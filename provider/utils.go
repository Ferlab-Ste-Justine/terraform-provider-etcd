package provider

func isStringInSlice(val string, slice []string) bool {
    for _, elem := range slice {
        if elem == val {
            return true
        }
    }
    return false
}