package openstack

import (
	"context"
	"fmt"
	"sync"

	"github.com/mobiledgex/edge-cloud/log"
)

var resourceLock sync.Mutex
var ReservedFloatingIPs map[string]string
var ReservedSubnets map[string]string

type ReservedResources struct {
	FloatingIpIds []string
	Subnets       []string
}

func (o *OpenstackPlatform) InitResourceReservations(ctx context.Context) {
	log.SpanLog(ctx, log.DebugLevelInfra, "InitResourceReservations")
	resourceLock.Lock()
	defer resourceLock.Unlock()
	ReservedFloatingIPs = make(map[string]string)
	ReservedSubnets = make(map[string]string)
}

// ReserveResources must be called from code that locks resourceLock
func (o *OpenstackPlatform) ReserveResources(ctx context.Context, resources *ReservedResources, reservedBy string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "ReserveResources", "resources", resources, "current fips", ReservedFloatingIPs, "current subnets", ReservedSubnets)

	var err error
	var fipsToCleanupOnErr []string
	var subnetsToCleanupOnErr []string

	for _, f := range resources.FloatingIpIds {
		err = o.reserveFloatingIP(ctx, f, reservedBy)
		if err != nil {
			break
		}
		fipsToCleanupOnErr = append(fipsToCleanupOnErr, f)
	}
	if err != nil {
		// Cleanup in case we reserved something and then hit an error
		for _, f := range fipsToCleanupOnErr {
			o.releaseFloatingIP(ctx, f)
		}
		return err
	}
	for _, s := range resources.Subnets {
		err := o.reserveSubnet(ctx, s, reservedBy)
		if err != nil {
			break
		}
		subnetsToCleanupOnErr = append(subnetsToCleanupOnErr, s)
	}
	if err != nil {
		// Cleanup in case we reserved something and then hit an error
		for _, s := range subnetsToCleanupOnErr {
			o.releaseSubnet(ctx, s)
		}
		return err
	}
	return nil
}

// ReleaseResources locks around resourceLock
func (o *OpenstackPlatform) ReleaseReservations(ctx context.Context, resources *ReservedResources) {
	log.SpanLog(ctx, log.DebugLevelInfra, "ReleaseReservations", "resources", resources, "current fips", ReservedFloatingIPs, "current subnets", ReservedSubnets)

	resourceLock.Lock()
	defer resourceLock.Unlock()
	for _, f := range resources.FloatingIpIds {
		o.releaseFloatingIP(ctx, f)
	}
	for _, s := range resources.Subnets {
		o.releaseSubnet(ctx, s)
	}
}

// reserveFloatingIP must be called from code that locks resourceLock
func (o *OpenstackPlatform) reserveFloatingIP(ctx context.Context, fipID string, reservedBy string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "ReserveFloatingIP", "fipID", fipID, "reservedBy", reservedBy)
	currUser, reserved := ReservedFloatingIPs[fipID]
	if reserved {
		return fmt.Errorf("Floating IP already reserved, fip: %s reservedBy: %s", fipID, currUser)
	}
	ReservedFloatingIPs[fipID] = reservedBy
	return nil
}

// reserveSubnet must be called from code that locks resourceLock
func (o *OpenstackPlatform) reserveSubnet(ctx context.Context, cidr string, reservedBy string) error {
	log.SpanLog(ctx, log.DebugLevelInfra, "reserveSubnet", "cidr", cidr, "reservedBy", reservedBy)
	currUser, reserved := ReservedSubnets[cidr]
	if reserved {
		return fmt.Errorf("Subnet CIDR already in reserved, cidr: %s reservedBy: %s", cidr, currUser)
	}
	ReservedSubnets[cidr] = reservedBy
	return nil
}

func (o *OpenstackPlatform) releaseFloatingIP(ctx context.Context, fipID string) {
	log.SpanLog(ctx, log.DebugLevelInfra, "releaseFloatingIP", "fipId", fipID)
	_, reserved := ReservedFloatingIPs[fipID]
	if !reserved {
		log.SpanLog(ctx, log.DebugLevelInfra, "Warning: Floating IP not reserved, cannot be released", "fipID", fipID)
	} else {
		delete(ReservedFloatingIPs, fipID)
	}
}

func (o *OpenstackPlatform) releaseSubnet(ctx context.Context, cidr string) {
	log.SpanLog(ctx, log.DebugLevelInfra, "releaseSubnet", "cidr", cidr)
	_, reserved := ReservedSubnets[cidr]
	if !reserved {
		log.SpanLog(ctx, log.DebugLevelInfra, "Warning: Subnet CIDR not reserved, cannot be released", "cidr", cidr)
	} else {
		delete(ReservedSubnets, cidr)
	}
}
