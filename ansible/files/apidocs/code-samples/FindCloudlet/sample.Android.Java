import com.mobiledgex.matchingengine.MatchingEngine;
FindCloudletRequest findCloudletRequest = createDefaultFindCloudletRequest(activityContext, location)
	.setCarrierName(carrierName)
	.build();
FindCloudletReply findCloudletReply = me.findCloudlet(findCloudletRequest, 10000);
