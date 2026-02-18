package utils




func Truncate(s string, maxLength int ) string {


    if len(s) <= maxLength {
        return s
    }

    return s[:maxLength] + "..." 
}
