//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Task struct {
	Command   string   `json:"command"`
	Inputs    []string `json:"inputs,omitempty"`
	Outputs   []string `json:"outputs,omitempty"`
	DependsOn []string `json:"depends_on,omitempty"`
}

type Config struct {
	DefaultTask string          `json:"default"`
	Tasks       map[string]Task `json:"tasks"`
}

func main() {
	// 1. Direct cycle
	cfgDirectCycle := Config{
		DefaultTask: "taskA",
		Tasks: map[string]Task{
			"taskA": {
				Command:   "echo 'taskA'",
				DependsOn: []string{"taskA"},
			},
		},
	}
	saveJSON("cycle_direct.json", cfgDirectCycle)

	// 2. Indirect cycle: A -> B -> C -> A
	cfgIndirectCycle := Config{
		DefaultTask: "taskA",
		Tasks: map[string]Task{
			"taskA": {Command: "echo 'A'", DependsOn: []string{"taskB"}},
			"taskB": {Command: "echo 'B'", DependsOn: []string{"taskC"}},
			"taskC": {Command: "echo 'C'", DependsOn: []string{"taskA"}},
		},
	}
	saveJSON("cycle_indirect.json", cfgIndirectCycle)

	// 3. Ghost dependency
	cfgGhost := Config{
		DefaultTask: "task1",
		Tasks: map[string]Task{
			"task1": {Command: "echo '1'", DependsOn: []string{"phantom_task"}},
		},
	}
	saveJSON("ghost_dep.json", cfgGhost)

	// 4. Ultra deep linear chain (50 tasks: T_00 -> T_01 -> ... -> T_49)
	tasksChain := make(map[string]Task)
	for i := 0; i < 50; i++ {
		name := fmt.Sprintf("chain_%02d", i)
		var deps []string
		if i > 0 {
			deps = []string{fmt.Sprintf("chain_%02d", i-1)}
		}
		tasksChain[name] = Task{
			Command:   fmt.Sprintf("cmd /c echo chain %02d", i),
			DependsOn: deps,
		}
	}
	saveJSON("deep_chain.json", Config{DefaultTask: "chain_49", Tasks: tasksChain})

	// 5. Wide Fan-out (20 workers -> 1 aggregator)
	tasksFanout := make(map[string]Task)
	var workerNames []string
	for i := 0; i < 20; i++ {
		wName := fmt.Sprintf("worker_%02d", i)
		workerNames = append(workerNames, wName)
		tasksFanout[wName] = Task{
			Command: fmt.Sprintf("cmd /c echo worker %02d finished", i),
		}
	}
	tasksFanout["aggregator"] = Task{
		Command:   "cmd /c echo All workers finished!",
		DependsOn: workerNames,
	}
	saveJSON("fanout_wide.json", Config{DefaultTask: "aggregator", Tasks: tasksFanout})

	// 6. Disconnected Multi-Island DAG
	tasksIsland := map[string]Task{
		"island1_a": {Command: "cmd /c echo island1_a"},
		"island1_b": {Command: "cmd /c echo island1_b", DependsOn: []string{"island1_a"}},
		"island2_x": {Command: "cmd /c echo island2_x"},
		"island2_y": {Command: "cmd /c echo island2_y", DependsOn: []string{"island2_x"}},
	}
	saveJSON("islands.json", Config{DefaultTask: "island1_b", Tasks: tasksIsland})

	fmt.Println("Generated all DAG stress fixtures successfully.")
}

func saveJSON(filename string, data interface{}) {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(filename, b, 0644); err != nil {
		panic(err)
	}
}
