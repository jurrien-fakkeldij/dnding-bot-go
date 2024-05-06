package features_test

import (
	"fmt"
	"jurrien/dnding-bot/features/steps"
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/spf13/pflag"
)

type stepCollection interface {
	InitializeSuite(suite *godog.TestSuiteContext) error
	InitializeScenario(scenario *godog.ScenarioContext) error
}

var opts = godog.Options{
	Output: colors.Colored(os.Stdout),
	Format: "progress", // can define default values
}

func init() {
	godog.BindCommandLineFlags("godog.", &opts) // godog v0.11.0 and later
}

func TestFeatures(t *testing.T) {
	pflag.Parse()
	opts.Paths = pflag.Args()
	commandSteps := &steps.CommandSteps{}
	stepCollections := []stepCollection{
		commandSteps,
	}
	suite := godog.TestSuite{
		Name: "DND-ING-BOT",
		ScenarioInitializer: func(scenario *godog.ScenarioContext) {
			fmt.Printf("Initializing Scenarios\n")
			for _, step := range stepCollections {
				if err := step.InitializeScenario(scenario); err != nil {
					panic(err)
				}
			}
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"."},
			TestingT: t, // Testing instance that will run subtests.
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
