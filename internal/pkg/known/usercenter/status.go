package usercenter

// Define user status.
const (
	// User has submitted registration information, the account is in a pending.
	// The user needs to complete email/phone number verification steps to transition to an active state.
	// The OneX project does not currently use this.
	UserStatusRegistered = "registered"
	// The user has registered and been verified, and can use the system normally.
	// Most user operations are performed in this state.
	UserStatusActived = "actived"
	// The user has entered the incorrect password too many times, and the account has been locked by the system.
	// The user needs to recover the password or contact the administrator to unlock the account.
	UserStatusLocked = "locked"
	// The user has been added to the system blacklist due to serious misconduct.
	// Blacklisted users cannot register new accounts or use the system.
	UserStatusBlacklisted = "blacklisted"
	// The administrator has manually disabled the user account, and the user cannot log in after being disabled.
	// This may be due to user misconduct or other reasons.
	UserStatusDisabled = "disabled"
	// The user has actively deleted their own account, or the administrator has deleted the user account.
	// The deleted account can be chosen to be soft-deleted (with some data retained) or completely deleted.
	UserStatusDeleted = "deleted"
)

// Define need status.
// We can directly update the database to the "Need" state to inform onex-nightwatch of what needs to be done.
// These statuses are only used for operation and maintenance purposes.
const (
	// UserStatusNeedActive informs onex-nightwatch that the user needs to be activated.
	UserStatusNeedActive = "need_active"
	// UserStatusNeedDisable informs onex-nightwatch that the user needs to be disabled.
	UserStatusNeedDisable = "need_disable"
)

// Define secret status.
const (
	SecretStatusDisabled = iota // Status used for disabling a secret.
	SecretStatusNormal          // Status used for enabling a secret.
)
