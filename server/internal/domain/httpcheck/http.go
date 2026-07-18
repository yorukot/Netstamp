package httpcheck

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
)

const (
	DefaultTimeoutMs      int32 = 10000
	MaxTimeoutMs          int32 = 60000
	MaxHeaders                  = 50
	MaxBodyBytes                = 64 * 1024
	MaxBodyContainsLength       = 1024
	MaxTargetLength             = 2048
	MaxRedirects                = 10
)

var (
	ErrInvalidConfig = errors.New("http config invalid")
	ErrInvalidResult = errors.New("http result invalid")
)

type Method string

const (
	MethodGet     Method = http.MethodGet
	MethodHead    Method = http.MethodHead
	MethodPost    Method = http.MethodPost
	MethodPut     Method = http.MethodPut
	MethodPatch   Method = http.MethodPatch
	MethodDelete  Method = http.MethodDelete
	MethodOptions Method = http.MethodOptions
)

type Status string

const (
	StatusSuccessful Status = "successful"
	StatusTimeout    Status = "timeout"
	StatusError      Status = "error"
)

type Header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Config struct {
	Method                Method                  `json:"method"`
	Headers               []Header                `json:"headers"`
	Body                  *string                 `json:"body,omitempty"`
	TimeoutMs             int32                   `json:"timeoutMs"`
	IPFamily              *domainnetwork.IPFamily `json:"ipFamily,omitempty"`
	FollowRedirects       bool                    `json:"followRedirects"`
	SkipTLSVerify         bool                    `json:"skipTlsVerify"`
	ExpectedStatusCodes   []int32                 `json:"-"`
	ExpectedStatusClasses []int32                 `json:"-"`
	BodyContains          *string                 `json:"bodyContains,omitempty"`
}

type StatusSelector struct {
	Kind  string  `json:"kind"`
	Code  *int32  `json:"code,omitempty"`
	Class *string `json:"class,omitempty"`
}

func (config Config) MarshalJSON() ([]byte, error) {
	type configAlias Config
	return json.Marshal(struct {
		configAlias
		ExpectedStatuses []StatusSelector `json:"expectedStatuses"`
	}{configAlias: configAlias(config), ExpectedStatuses: config.StatusSelectors()})
}

func (config *Config) UnmarshalJSON(data []byte) error {
	type configAlias Config
	decoded := struct {
		configAlias
		ExpectedStatuses []StatusSelector `json:"expectedStatuses"`
	}{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*config = Config(decoded.configAlias)
	for _, selector := range decoded.ExpectedStatuses {
		switch selector.Kind {
		case "code":
			if selector.Code == nil || selector.Class != nil {
				return errors.New("http status code selector must include only code")
			}
			config.ExpectedStatusCodes = append(config.ExpectedStatusCodes, *selector.Code)
		case "class":
			if selector.Class == nil || selector.Code != nil || !validStatusClass(*selector.Class) {
				return errors.New("http status class selector is invalid")
			}
			config.ExpectedStatusClasses = append(config.ExpectedStatusClasses, int32((*selector.Class)[0]-'0'))
		default:
			return errors.New("http status selector kind is invalid")
		}
	}
	var err error
	config.ExpectedStatusCodes, config.ExpectedStatusClasses, err = VNExpectedStatuses(config.ExpectedStatusCodes, config.ExpectedStatusClasses)
	return err
}

func validStatusClass(value string) bool {
	return len(value) == 3 && value[0] >= '1' && value[0] <= '5' && value[1:] == "xx"
}

func (config Config) StatusSelectors() []StatusSelector {
	selectors := make([]StatusSelector, 0, len(config.ExpectedStatusClasses)+len(config.ExpectedStatusCodes))
	for _, class := range config.ExpectedStatusClasses {
		value := strconv.Itoa(int(class)) + "xx"
		selectors = append(selectors, StatusSelector{Kind: "class", Class: &value})
	}
	for _, code := range config.ExpectedStatusCodes {
		value := code
		selectors = append(selectors, StatusSelector{Kind: "code", Code: &value})
	}
	return selectors
}

func RedactTarget(value string) string {
	parsed, err := url.Parse(value)
	if err != nil {
		return ""
	}
	parsed.RawQuery = ""
	parsed.ForceQuery = false
	parsed.Fragment = ""
	parsed.User = nil
	return parsed.String()
}

func DefaultConfig() Config {
	return Config{
		Method:                MethodGet,
		Headers:               []Header{},
		TimeoutMs:             DefaultTimeoutMs,
		FollowRedirects:       true,
		ExpectedStatusCodes:   []int32{},
		ExpectedStatusClasses: []int32{2, 3},
	}
}

func VNTarget(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", errors.New("must be provided")
	}
	if len(value) > MaxTargetLength {
		return "", fmt.Errorf("must be at most %d characters", MaxTargetLength)
	}
	parsed, err := url.ParseRequestURI(value)
	if err != nil {
		return "", errors.New("must be a valid HTTP URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", errors.New("scheme must be http or https")
	}
	if parsed.Hostname() == "" {
		return "", errors.New("host must be provided")
	}
	if parsed.User != nil {
		return "", errors.New("userinfo is not allowed")
	}
	if parsed.Fragment != "" {
		return "", errors.New("fragment is not allowed")
	}
	return parsed.String(), nil
}

func VNMethod(value Method) (Method, error) {
	method := Method(strings.ToUpper(strings.TrimSpace(string(value))))
	switch method {
	case MethodGet, MethodHead, MethodPost, MethodPut, MethodPatch, MethodDelete, MethodOptions:
		return method, nil
	default:
		return "", errors.New("invalid HTTP method")
	}
}

func VNHeaders(values []Header) ([]Header, error) {
	if len(values) > MaxHeaders {
		return nil, fmt.Errorf("must contain at most %d headers", MaxHeaders)
	}
	output := make([]Header, len(values))
	for i, value := range values {
		name := http.CanonicalHeaderKey(strings.TrimSpace(value.Name))
		if name == "" || !validHeaderName(name) {
			return nil, fmt.Errorf("header %d has an invalid name", i)
		}
		if strings.ContainsAny(value.Value, "\r\n") {
			return nil, fmt.Errorf("header %d has an invalid value", i)
		}
		output[i] = Header{Name: name, Value: value.Value}
	}
	return output, nil
}

func validHeaderName(value string) bool {
	for _, r := range value {
		if r <= 32 || r >= 127 || strings.ContainsRune("()<>@,;:\\\"/[]?={}\t", r) {
			return false
		}
	}
	return true
}

func VNBody(method Method, body *string) (*string, error) {
	if body == nil {
		return nil, nil //nolint:nilnil // Nil means no request body.
	}
	if method == MethodGet || method == MethodHead {
		return nil, errors.New("GET and HEAD requests cannot include a body")
	}
	if len([]byte(*body)) > MaxBodyBytes {
		return nil, fmt.Errorf("must be at most %d bytes", MaxBodyBytes)
	}
	value := *body
	return &value, nil
}

func VNTimeoutMs(value int32) (int32, error) {
	if value < 1 || value > MaxTimeoutMs {
		return 0, fmt.Errorf("must be between 1 and %d", MaxTimeoutMs)
	}
	return value, nil
}

func VNIPFamily(value *domainnetwork.IPFamily) (*domainnetwork.IPFamily, error) {
	if value == nil {
		return nil, nil //nolint:nilnil // Nil means automatic address-family selection.
	}
	if *value != domainnetwork.IPFamilyInet && *value != domainnetwork.IPFamilyInet6 {
		return nil, errors.New("invalid IP family")
	}
	return value, nil
}

func VNExpectedStatuses(codes, classes []int32) ([]int32, []int32, error) {
	if len(codes) == 0 && len(classes) == 0 {
		return nil, nil, errors.New("at least one expected status is required")
	}
	for _, code := range codes {
		if code < 100 || code > 599 {
			return nil, nil, fmt.Errorf("status code %d must be between 100 and 599", code)
		}
	}
	for _, class := range classes {
		if class < 1 || class > 5 {
			return nil, nil, fmt.Errorf("status class %dxx is invalid", class)
		}
	}
	codes = sortedUnique(codes)
	classes = sortedUnique(classes)
	return codes, classes, nil
}

func sortedUnique(values []int32) []int32 {
	output := append([]int32(nil), values...)
	slices.Sort(output)
	return slices.Compact(output)
}

func VNBodyContains(value *string) (*string, error) {
	if value == nil {
		return nil, nil //nolint:nilnil // Nil disables response-body matching.
	}
	if *value == "" {
		return nil, errors.New("must not be empty")
	}
	if len(*value) > MaxBodyContainsLength {
		return nil, fmt.Errorf("must be at most %d characters", MaxBodyContainsLength)
	}
	normalized := *value
	return &normalized, nil
}

func MatchesStatus(config Config, code int) bool {
	for _, expected := range config.ExpectedStatusCodes {
		if code == int(expected) {
			return true
		}
	}
	class := code / 100
	return slices.ContainsFunc(config.ExpectedStatusClasses, func(expected int32) bool {
		return class == int(expected)
	})
}

type Result struct {
	StartedAt            time.Time
	FinishedAt           time.Time
	DurationMs           int32
	Status               Status
	DNSDurationMs        *float64
	ConnectDurationMs    *float64
	TLSDurationMs        *float64
	TTFBDurationMs       *float64
	ResolvedIP           *netip.Addr
	IPFamily             *domainnetwork.IPFamily
	StatusCode           *int32
	FinalURL             *string
	RedirectCount        int32
	ResponseBytes        *int64
	ResponseTruncated    bool
	BodyMatched          *bool
	TLSVersion           *string
	TLSCipherSuite       *string
	CertificateNotBefore *time.Time
	CertificateNotAfter  *time.Time
	ErrorCode            *string
	ErrorMessage         *string
}

type ResultStorageInput struct {
	ProbeStorageID int64
	CheckStorageID int64
	Result
}

type LatestResultQuery struct {
	ProjectID string
	ProbeID   string
	CheckID   string
}

type LatestResult struct {
	ProbeID string
	CheckID string
	Result  Result
}

type LatestResultList struct {
	Results []LatestResult
}

type (
	SeriesReadMode   string
	SeriesSource     string
	SeriesResolution string
)

const (
	SeriesReadModeRaw         SeriesReadMode   = "raw"
	SeriesReadModeBucket      SeriesReadMode   = "bucket"
	SeriesReadModeRollup      SeriesReadMode   = "rollup"
	SeriesSourceRaw           SeriesSource     = "raw"
	SeriesSourceAggregate     SeriesSource     = "aggregate"
	SeriesResolutionRaw       SeriesResolution = "raw"
	SeriesResolutionBucket    SeriesResolution = "bucket"
	SeriesResolutionOneMinute SeriesResolution = "1m"
)

type SeriesPointCountQuery struct {
	ProjectID, ProbeID, CheckID string
	From, To                    time.Time
}
type SeriesReadQuery struct {
	ProjectID, ProbeID, CheckID string
	From, To                    time.Time
	Series                      []string
	MaxDataPoints               int32
	Mode                        SeriesReadMode
}
type SeriesReadPlan struct {
	Mode        SeriesReadMode
	Source      SeriesSource
	Resolution  SeriesResolution
	TotalPoints int64
}
type (
	SeriesData  struct{ Points []SeriesPoint }
	SeriesPoint struct {
		Timestamp time.Time
		Value     float64
	}
)

type InsightSummaryQuery struct {
	ProjectID, ProbeID, CheckID string
	From, To                    time.Time
	Source                      SeriesSource
}
type InsightSummary struct {
	TotalResults                                          int64
	AverageTotalMs, MaxTotalMs, AverageTTFBMs, MaxTTFBMs  *float64
	FailurePercent, SuccessRate, CertificateDaysRemaining *float64
	Samples                                               int64
}
