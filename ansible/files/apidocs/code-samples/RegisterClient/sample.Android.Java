import com.mobiledgex.matchingengine.MatchingEngine;

AppClient.Tag tag = AppClient.Tag.newBuilder()
	.setType(key)
	.setData(datavalue);
RegisterClientRequest registerClientRequest = createDefaultRegisterClientRequest(activityContext, developerName)
	.setAuthToken(authToken)
	.setTags(0, tag)
	.build();
me.registerClient(registerClientRequest, 10000);
