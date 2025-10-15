package namespace

import "github.com/google/go-cmp/cmp"

// LateInitializeStringPtr returns `from` if `in` is nil and `from` is non-empty,
// in other cases it returns `in`.
func LateInitializeStringPtr(in *string, from string) *string {
	if in == nil && from != "" {
		return &from
	}
	return in
}

// LateInitializeInt64Ptr returns `from` if `in` is nil and `from` is non-empty,
// in other cases it returns `in`.
func LateInitializeInt64Ptr(in *int64, from int64) *int64 {
	if in == nil && from != 0 {
		return &from
	}
	return in
}

// StringToPtr converts string to *string
func StringToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// StringValue converts a *string to string
func StringValue(ptr *string) string {
	if ptr != nil {
		return *ptr
	}
	return ""
}

// Int64Value converts a *int64 to int64
func Int64Value(ptr *int64) int64 {
	if ptr != nil {
		return *ptr
	}
	return 0
}

// BoolValue converts a *bool to bool
func BoolValue(ptr *bool) bool {
	if ptr != nil {
		return *ptr
	}
	return false
}

// IsBoolEqualToBoolPtr compares a *bool with bool
func IsBoolEqualToBoolPtr(bp *bool, b bool) bool {
	if bp != nil {
		if !cmp.Equal(*bp, b) {
			return false
		}
	}
	return true
}

// IsInt64EqualToInt64Ptr compares a *bool with bool
func IsInt64EqualToInt64Ptr(ip *int64, i int64) bool {
	if ip != nil {
		if !cmp.Equal(*ip, i) {
			return false
		}
	}
	return true
}
