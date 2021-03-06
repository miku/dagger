package dagger

import (
	"fmt"
	"log"

	"github.com/fatih/structs"
)

func TaskDeps(r Outputter) map[Outputter][]Outputter {
	var result = make(map[Outputter][]Outputter)
	var queue = []Outputter{r}

	for len(queue) > 0 {
		r := queue[0]
		queue = queue[1:]
		if rr, ok := r.(Requirer); ok {
			for _, v := range rr.Requires() {
				queue = append(queue, v)
				result[r] = append(result[r], v)
			}
		}
	}
	return result
}

func TopoSort(m map[Outputter][]Outputter) []Outputter {
	var order []Outputter
	seen := make(map[Outputter]bool)
	var visitAll func(items []Outputter)
	visitAll = func(items []Outputter) {
		for _, item := range items {
			if !seen[item] {
				seen[item] = true
				visitAll(m[item])
				order = append(order, item)
			}
		}
	}

	var keys []Outputter
	for key := range m {
		keys = append(keys, key)
	}
	visitAll(keys)
	return order
}

func InitializeTask(runner *Runner) error {
	s := structs.New(*runner)
	for _, _ = range s.Fields() {
		// apply defaults here
		// log.Println(f, f.Tag("default"))
	}
	return nil
}

func Build(outputter Outputter) error {
	prereqs := TaskDeps(outputter)

	for _, o := range TopoSort(prereqs) {
		log.Printf("%# v - %# v - %v", o, o.Output(), o.Output().Exists())
		if !o.Output().Exists() {
			log.Printf("Running %# v...", o)
			if rr, ok := o.(Runner); ok {
				// TODO(miku): apply defaults to rr here
				if err := rr.Run(); err != nil {
					return err
				}
			} else {
				return fmt.Errorf("cannot create missing task output for: %# v", o)
			}
		}
	}
	return nil
}
