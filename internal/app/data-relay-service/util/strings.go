package util

// StrNotCurrentlyLoggedIn : Error msg string, User's request is not authorized.
const StrNotCurrentlyLoggedIn = "Not currently authorized."

// StrIncorrectParametersPassed : Error msg string, Wrong parameter passed.
const StrIncorrectParametersPassed = "Parameters passed are not valid."

// StrUserNotFoundError : Error msg string, User not found.
const StrUserNotFoundError = "Username / Password combination does not exist."

// StrTemplateNotFoundError : Error msg string, Template not found.
const StrTemplateNotFoundError = "Template not found -> "

// StrStreamSendError : Error msg string, Stream send failed.
const StrStreamSendError = "Error with Stream: %v - Error: %v"

// StrDatabaseConnectError : Error msg string, Database failed to connect.
const StrDatabaseConnectError = "Error when connecting to database."

// StrInvalidInputError : Error msg string, Bad input attempt.
const StrInvalidInputError = "Invalid input. Please try again."

// StrConfigLoadError : Error msg string, Config not found.
const StrConfigLoadError = "Failed to load app configuration; %v"

// StrDatabaseConnected : Info msg string, Database connected.
const StrDatabaseConnected = "Successfully connected to database."

// StrRedisConnectError : Error msg string, Redis failed to connect
const StrRedisConnectError = "Failed to connect to redis server."

// StrRedisConnectSuccess : Info msg string, Redis successful connect
const StrRedisConnectSuccess = "Successfully connected to redis server."

// StrDiscoveryRegFail : Error msg string, Discovery Service failed to register.
const StrDiscoveryRegFail = "Failed to register to discovery service."

// StrDiscoveryRegSuccess : Info msg string, Discovery Service successfully registered.
const StrDiscoveryRegSuccess = "Successfully registered with discovery service."

// StrDiscoveryUnRegFail : Error msg string, Discovery Service unsuccessful unregistered.
const StrDiscoveryUnRegFail = "Failed to unregister with discovery service"

// StrDiscoveryUnRegSuccess : Info msg string, Discovery Service successfully unregistered.
const StrDiscoveryUnRegSuccess = "UnRegistered with discovery service"

// StrDiscoveryCheckFail : Error msg string, Discovery Service checkin failure.
const StrDiscoveryCheckFail = "CheckIn with discovery service failed."

// StrDiscoveryCheckSuccess : Info msg string, Discovery Service checking success.
const StrDiscoveryCheckSuccess = "Succesful CheckIn with discovery service"
