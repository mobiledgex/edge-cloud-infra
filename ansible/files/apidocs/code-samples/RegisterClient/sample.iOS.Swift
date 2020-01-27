import IOSMatchingEngine
matchingEngine = MatchingEngine()
let registerClientRequest = matchingEngine.createRegisterClientRequest(devName: devName, appName: appName, appVers: appVers, carrierName: carrierName, authToken: authToken)
let registerClientReplyPromise = matchingEngine.registerClient(request: registerClientRequest)
	.then { registerClientReply in
	\* handle registerClientReply*\
	}
