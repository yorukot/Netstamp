# Domain Layer

The domain layer is where stable domain models live.

Domain packages own domain-level validation and normalization. Functions prefixed with `VN` validate and normalize input before it moves deeper into the system.

Domain packages may import the Go standard library, third-party libraries, and other `internal/domain/...` packages when the model relationship requires it. Domain packages must not import non-domain Netstamp packages such as controller, application, transport, infrastructure, platform, config, logger, database, generated, or runtime packages.

# TODO

RIght now we have a lot of duplicate VN on the domain layer we should update it to use the global value
