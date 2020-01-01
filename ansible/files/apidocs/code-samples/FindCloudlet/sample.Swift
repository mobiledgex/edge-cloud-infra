import IOSMatchingEngine
matchingEngine = MatchingEngine()
let findCloudletRequest = matchingEngine.createFindCloudletRequest(carrierName: carrierName, gpsLocation: loc, devName: devName, appName: appName, appVers: appVers)
let findCloudletReplyPromise = matchingEngine.findCloudlet(request: findCloudletRequest)
	.then { findCloudletReply in
	\* handle findCloudletReply*\
	}
