package legacy

// Inject legacy variable into context
func InjectLegacyVar(ctx map[string]string, serviceType string) map[string]string {
	res := map[string]string{
		"notificationType": serviceType,
	}
	for k, v := range ctx {
		res[k] = v
	}
	return res
}
