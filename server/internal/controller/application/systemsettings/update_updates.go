package systemsettings

import "context"

func (s *Service) UpdateUpdates(ctx context.Context, input UpdateUpdatesInput) (UpdatesSettings, error) {
	ctx, flow := s.startSettingsFlow(ctx, "update", ResourceUpdates, input.CurrentUserID)
	defer flow.end()
	if authorizeErr := flow.authorize(ctx); authorizeErr != nil {
		return UpdatesSettings{}, authorizeErr
	}

	var result UpdatesSettings
	txErr := s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		if lockErr := s.repo.LockSystemSettingsResource(txCtx, string(ResourceUpdates)); lockErr != nil {
			return lockErr
		}
		current, loadErr := s.EffectiveUpdates(txCtx)
		if loadErr != nil {
			return loadErr
		}
		next := current
		if input.CheckForUpdates != nil {
			next.CheckForUpdates = *input.CheckForUpdates
		}
		if current == next {
			result = current
			return nil
		}
		if persistErr := s.upsertPublic(txCtx, keyUpdateCheckEnabled, next.CheckForUpdates, &input.CurrentUserID); persistErr != nil {
			return persistErr
		}
		result = next
		return nil
	})
	return result, txErr
}
