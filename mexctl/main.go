package main

import (
	"flag"
	"fmt"
	"os"
	"reflect"

	sh "github.com/codeskyblue/go-sh"
	"github.com/mobiledgex/edge-cloud-infra/mexos"
	"github.com/mobiledgex/edge-cloud/log"
)

var clusterOps = map[string]func(*mexos.Manifest) error{
	"create": mexos.MEXClusterCreateManifest,
	"remove": mexos.MEXClusterRemoveManifest,
}

var platformOps = map[string]func(*mexos.Manifest) error{
	"create": mexos.MEXPlatformInitManifest,
	"remove": mexos.MEXPlatformCleanManifest,
}

var applicationOps = map[string]func(*mexos.Manifest) error{
	"create": mexos.MEXAppCreateAppManifest,
	"remove": mexos.MEXAppDeleteAppManifest,
}

var openstackOps = map[string]func(*mexos.Manifest) error{}

var categories = map[string]map[string]func(*mexos.Manifest) error{
	"cluster":     clusterOps,
	"platform":    platformOps,
	"application": applicationOps,
	"openstack":   openstackOps,
}

var mainflag = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

func printUsage() {
	originalUsage()
	fmt.Println("mex -stack myvals.yaml {platform|cluster|application} {create|remove}")
	fmt.Println("mex -stack myvals.yaml openstack ...")
}

var originalUsage func()

func main() {
	var err error
	help := mainflag.Bool("help", false, "help")
	debugLevels := mainflag.String("d", "", fmt.Sprintf("comma separated list of %v", log.DebugLevelStrings))
	base := mainflag.String("base", ".", "base containing templates")
	stack := mainflag.String("stack", "", "stack values")
	originalUsage = mainflag.Usage
	mainflag.Usage = printUsage
	if err = mainflag.Parse(os.Args[1:]); err != nil {
		log.FatalLog("parse error", "error", err)
	}
	if *help {
		printUsage()
		os.Exit(0)
	}
	log.SetDebugLevelStrs(*debugLevels)
	//XXX TODO make log to a remote server / aggregator
	args := mainflag.Args()
	if len(args) < 2 {
		printUsage()
		fmt.Println("insufficient args")
		os.Exit(1)
	}
	_, ok := categories[args[0]]
	if !ok {
		printUsage()
		fmt.Println("valid categories are", "categories", reflect.ValueOf(categories).MapKeys())
		os.Exit(1)
	}
	if *stack == "" {
		printUsage()
		fmt.Println("missing stack")
		os.Exit(1)
	}
	if len(args) < 2 {
		printUsage()
		fmt.Println("insufficient args")
		os.Exit(1)
	}
	log.DebugLog(log.DebugLevelMexos, "getting mf from stack", "file", *stack)
	mf := &mexos.Manifest{}
	if err := mexos.GetVaultEnv(mf, *stack); err != nil {
		log.FatalLog("cannot get mf", "uri", *stack, "error", err)
	}
	kind := args[0]
	if err := mexos.FillManifest(mf, kind, *base); err != nil {
		log.FatalLog("cannot fill manifest", "error", err, "kind", kind, "base", *base)
	}
	if err := mexos.CheckManifest(mf); err != nil {
		log.FatalLog("incorrect manifest", "error", err)
	}
	if _, err := mexos.NewRootLBManifest(mf); err != nil {
		log.FatalLog("can't get new rootLB", "error", err)
	}
	if err := mexos.MEXInit(mf); err != nil {
		log.FatalLog("cannot init mex", "error", err)
	}
	ops := args[1:]
	log.DebugLog(log.DebugLevelMexos, "call", "kind", kind, "ops", ops)
	err = callOps(mf, kind, ops...)
	if err != nil {
		log.FatalLog("ops failure", "kind", kind, "ops", ops, "error", err)
	}
	os.Exit(0)
}

func callOps(mf *mexos.Manifest, kind string, ops ...string) error {
	if kind == "openstack" {
		vs := make([]interface{}, len(ops))
		for i, v := range ops {
			vs[i] = v
		}
		out, err := sh.Command("openstack", vs...).Output()
		if err != nil {
			return fmt.Errorf("error, openstack %v, %v", ops, err)
		}
		fmt.Println(string(out))
		return nil
	}
	_, ok := categories[kind]
	if !ok {
		return fmt.Errorf("invalid category %s", kind)
	}
	op := ops[0]
	err := categories[kind][op](mf)
	if err != nil {
		return err
	}
	return nil
}
