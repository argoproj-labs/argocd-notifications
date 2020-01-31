package expr

var helpers map[string]interface{}

func init() {
	helpers = make(map[string]interface{})
}

func register(namespace string, entry map[string]interface{}) {
	helpers[namespace] = entry
}

func Spawn() map[string]interface{} {
	clone := make(map[string]interface{})
	for namespace, helper := range helpers {
		clone[namespace] = helper
	}

	return clone
}
