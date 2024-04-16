package features_test

import (
	"fmt"
	"jurrien/dnding-bot/features/steps"
	"testing"

	"github.com/cucumber/godog"
)

type stepCollection interface {
	InitializeSuite(suite *godog.TestSuiteContext) error
	InitializeScenario(scenario *godog.ScenarioContext) error
}

func TestFeatures(t *testing.T) {
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
