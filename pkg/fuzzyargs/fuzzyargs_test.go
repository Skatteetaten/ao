package fuzzyargs

import (
	"testing"

	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/skatteetaten/ao/pkg/serverapi"
)

func TestGetOneFile(t *testing.T) {
	fuzzyArgs, err := initiateFuzzyArgs()

	const argument = "utv0/afs-import.json"
	const expected = "utv0/afs-import.json"
	err = fuzzyArgs.PopulateFuzzyFile(getArgs(argument))
	if err != nil {
		t.Errorf("Error in PopulateFuzzyFile(%v): %v", argument, err.Error())
	}
	filename, err := fuzzyArgs.GetFile()
	if err != nil {
		t.Errorf("Error in GetFile(%v): %v", argument, err.Error())
	} else {
		if filename != expected {
			t.Errorf("Eror in GetFile(%v): Expected %v, got %v", argument, expected, filename)
		}
	}
}

func TestGetOneFileArray(t *testing.T) {
	fuzzyArgs, err := initiateFuzzyArgs()

	const argument1 = "utv0"
	const argument2 = "afs-import.json"
	const expected = "utv0/afs-import.json"
	err = fuzzyArgs.PopulateFuzzyFile(getArgs(argument1, argument2))
	if err != nil {
		t.Errorf("Error in PopulateFuzzyFile(%v %v):  %v", argument1, argument2, err.Error())
	} else {
		filename, err := fuzzyArgs.GetFile()
		if err != nil {
			t.Errorf("Error in GetFile(%v): %v %v", argument1, argument2, err.Error())
		} else {
			if filename != expected {
				t.Errorf("Eror in GetFile(%v %v): Expected %v, got %v", argument1, argument2, expected, filename)
			}
		}
	}
}

func TestGetOneFuzzyFile(t *testing.T) {
	fuzzyArgs, err := initiateFuzzyArgs()

	const argument = "v0/afs"
	const expected = "utv0/afs-import.json"
	err = fuzzyArgs.PopulateFuzzyFile(getArgs(expected))
	if err != nil {
		t.Errorf("Error in PopulateFuzzyFile(%v): %v", expected, err.Error())
	} else {

		filename, err := fuzzyArgs.GetFile()
		if err != nil {
			t.Errorf("Error in GetFile(%v): %v", expected, err.Error())
		} else {
			if filename != expected {
				t.Errorf("Eror in GetFile(%v): Expected %v, got %v", argument, expected, filename)
			}
		}
	}
}

func TestGetNonUniqueEnvFuzzyFile(t *testing.T) {
	fuzzyArgs, err := initiateFuzzyArgs()

	const expected = "0/afs-import.json"
	err = fuzzyArgs.PopulateFuzzyFile(getArgs(expected))
	if err == nil {
		t.Errorf("Error in PopulateFuzzyFile(%v): Expected duplicate error", err.Error())
	}
}

func TestGetNonUniqueAppFuzzyFile(t *testing.T) {
	fuzzyArgs, err := initiateFuzzyArgs()

	const expected = "v0/im"
	err = fuzzyArgs.PopulateFuzzyFile(getArgs(expected))
	if err == nil {
		t.Errorf("Error in PopulateFuzzyFile(%v): Expected duplicate error", expected)
	}
}

func TestGetUniqueApp(t *testing.T) {
	fuzzyArgs, err := initiateFuzzyArgs()
	const argument = "bas-dev/console"
	err = fuzzyArgs.PopulateFuzzyEnvAppList(getArgs(argument), false)
	if err != nil {
		t.Errorf("Error in PopulateFuzzyEnvAppList(%v): %v", argument, err.Error())
	}
}

func TestGetOneFuzzyApp(t *testing.T) {
	fuzzyArgs, err := initiateFuzzyArgs()
	const argument = "bas/con"
	const expectedApp = "console"
	const expectedEnv = "bas-dev"
	err = fuzzyArgs.PopulateFuzzyEnvAppList(getArgs(argument), false)
	if err != nil {
		t.Errorf("Error in PopulateFuzzyEnvAppList(%v): %v", argument, err.Error())
	} else {
		app, err := fuzzyArgs.GetApp()
		if err != nil {
			t.Errorf("Error in GetApp: %v", err.Error())
		} else {
			if app != expectedApp {
				t.Errorf("Error in TestGetOneFuzzyApp, Expected app %v, got %v", expectedApp, app)
			}
		}
		env, err := fuzzyArgs.GetEnv()
		if err != nil {
			t.Errorf("Error in GetEnv: %v", err.Error())
		} else {
			if env != expectedEnv {
				t.Errorf("Error in TestGetOneFuzzyApp, Expected env %v, got %v", expectedEnv, env)
			}
		}

	}
}

func TestGetFuzzyApp(t *testing.T) {
	fuzzyArgs, err := initiateFuzzyArgs()
	const argument = "con"
	const expected = "console"
	app, err := fuzzyArgs.GetFuzzyApp(argument)
	if err != nil {
		t.Errorf("Error in GetFuzzyApp: %v", err.Error())
	} else {
		if app != expected {
			t.Errorf("Error in TestGetFuzzyApp: Expected %v, got %v", expected, app)
		}
	}
}

func TestGetFuzzyEnv(t *testing.T) {
	fuzzyArgs, err := initiateFuzzyArgs()
	const argument = "bas-d"
	const expected = "bas-dev"
	env, err := fuzzyArgs.GetFuzzyEnv(argument)
	if err != nil {
		t.Errorf("Error in GetFuzzyEnv: %v", err.Error())
	} else {
		if env != expected {
			t.Errorf("Error in TestGetFuzzyEnv: Expected %v, got %v", expected, env)
		}
	}
}

func TestApp2File(t *testing.T) {
	fuzzyArgs, err := initiateFuzzyArgs()

	const argument = "s-dev/cons"
	const expected = "bas-dev/console.json"
	filename, err := fuzzyArgs.App2File(argument)
	if err != nil {
		t.Errorf("Error in TestApp2File: %v", err.Error())
	} else {
		if filename != expected {
			t.Errorf("Error in TestApp2File: Expected %v, got %v", expected, filename)
		}
	}
}

func initiateFuzzyArgs() (fuzzyArgs *FuzzyArgs, err error) {
	config := configuration.NewTestConfiguration()
	request := auroraconfig.GetAuroraConfigRequest(config)
	response, err := serverapi.CallApiWithRequest(request, config)
	if err != nil {
		return fuzzyArgs, err
	}
	auroraConfig, err := auroraconfig.Response2AuroraConfig(response)
	if err != nil {
		return fuzzyArgs, err
	}

	fuzzyArgs = new(FuzzyArgs)
	fuzzyArgs.Init(&auroraConfig)
	return fuzzyArgs, nil
}

func getArgs(argN ...string) (args []string) {
	return argN
}
