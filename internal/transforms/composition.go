// Copyright 2020 Enea Guidi (hmny). All rights reserved.
// This file are distributed under the General Public License v 3.0.
// A copy of abovesaid license can be found in the LICENSE file.

// Package transforms declares the types and functions used to transform and work with some type of FSA.
// Come of the transformation implemented here are standard such as determinization (Subset Construction),
// minimization but more are specifically related to Choreia (GoroutineFSA extraction & Composition)
//
package transforms

import (
	"fmt"
	"log"

	list "github.com/emirpasic/gods/lists/singlylinkedlist"
	"github.com/its-hmny/Choreia/internal/data_structures/fsa"
)

type SimulationFSA struct {
	GoroutineFSA
	currentState int
}

var (
	// ! simDiamond            = (*SimulationDiamond)(set.New())
	simDiamond            = list.New()
	choreographyAutomaton = fsa.New()
)

func ComposeGoroutines(goroutines map[string]GoroutineFSA) *fsa.FSA {
	defer func() {
		// simDiamond = (*SimulationDiamond)(set.New())
		simDiamond = list.New()
		choreographyAutomaton = fsa.New()
	}()

	mainGrFSA, exist := goroutines["main (0)"]
	entrypoint := list.New(SimulationFSA{mainGrFSA, 0})

	if !exist {
		log.Fatal("Could not find GoroutineFSA for 'main'")
	}

	explore(entrypoint, goroutines)
	return choreographyAutomaton
}

func explore(sim *list.List, goroutines map[string]GoroutineFSA) {

	for indexA, itemA := range sim.Values() {
		participantA := itemA.(SimulationFSA)

		// Unary transitions handling (Spawn)
		participantA.Automaton.ForEachTransition(func(from, to int, t fsa.Transition) {
			// Only interested in unary transitions from the current state
			if from != participantA.currentState || t.Move != fsa.Spawn {
				return
			}

			// Makes a copy of the current participant adn updates its current state
			copy := participantA
			copy.currentState = to
			// Creates a new SimulationFSA with the new participant state instead of the old one
			newSim := list.New(sim.Values()...)

			// TODO
			pAIndex, _ := newSim.Find(func(index int, item interface{}) bool {
				current := item.(SimulationFSA)
				return participantA.Name == current.Name
			})
			// TODO
			newSim.Remove(pAIndex)
			newSim.Insert(pAIndex, copy)

			grFSA, exist := goroutines[t.Label]
			if !exist {
				log.Fatalf("Could not find GoroutineFSA for %s", t.Label)
			}
			newSim.Add(SimulationFSA{grFSA, 0})

			// If newSim isn't already contained in the simulation diamond then is added and a
			// recursive call on this newly found system configuration is done to explore its subgraph
			if !simDiamond.Contains(newSim) {
				simDiamond.Add(newSim)
				explore(newSim, goroutines)
			}

			// Retrieve the node id for the prev simulation state in the automaton
			oldSimId, _ := simDiamond.Find(func(index int, item interface{}) bool {
				current := item.(*list.List)
				return current.Contains(sim.Values()...) && sim.Contains(current.Values()...)
			})
			// Retrieve the node id for the new simulation state in the automaton
			newSimId, _ := simDiamond.Find(func(index int, item interface{}) bool {
				current := item.(*list.List)
				return current.Contains(sim.Values()...) && sim.Contains(current.Values()...)
			})

			// Adds a new edge in the choreography automaton
			newT := fsa.Transition{
				Move:  fsa.Empty,
				Label: fmt.Sprintf("%s \u22C1 %s", participantA.Name, t.Label),
			}
			choreographyAutomaton.AddTransition(oldSimId, newSimId, newT)
			fmt.Println(newT)
		})

		// Binary transitions handling (Send, Recv)
		for indexB, itemB := range sim.Values() {
			participantB := itemB.(SimulationFSA)

			participantA.Automaton.ForEachTransition(func(fromA, toA int, tA fsa.Transition) {
				participantB.Automaton.ForEachTransition(func(fromB, toB int, tB fsa.Transition) {
					// Makes a copy of both A and B and updates their respective "currentstate" fields
					copyA, copyB := participantA, participantB
					copyA.currentState, copyB.currentState = toA, toB

					// Creates a new SimulationFSA with the new participant state instead of the old one
					newSim := list.New(sim.Values()...)
					newSim.Remove(indexA)
					newSim.Insert(indexA, copyA)
					newSim.Remove(indexB)
					newSim.Insert(indexB, copyB)

					// // TODO
					// pBIndex, _ := newSim.Find(func(index int, item interface{}) bool {
					// 	current := item.(SimulationFSA)
					// 	return participantB.Name == current.Name
					// })
					// // TODO
					// newSim.Remove(pBIndex)
					// newSim.Insert(pBIndex, copyA)

					// Retrieve the node id for the prev simulation state in the automaton
					oldSimId, _ := simDiamond.Find(func(index int, item interface{}) bool {
						current := item.(*list.List)
						return current.Contains(sim.Values()...) && sim.Contains(current.Values()...)
					})
					// Retrieve the node id for the new simulation state in the automaton
					newSimId, _ := simDiamond.Find(func(index int, item interface{}) bool {
						current := item.(*list.List)
						return current.Contains(sim.Values()...) && sim.Contains(current.Values()...)
					})

					if tA.Move == fsa.Send && tB.Move == fsa.Recv && tA.Label == tB.Label {
						// Adds a new edge in the choreography automaton
						newT := fsa.Transition{
							Move:  fsa.Empty,
							Label: fmt.Sprintf("%s \u2192 %s", participantA.Name, participantB.Name),
						}
						fmt.Println(newT)
						choreographyAutomaton.AddTransition(oldSimId, newSimId, newT)

						// If newSim isn't already contained in the simulation diamond then is added and a
						// recursive call on this newly found system configuration is done to explore
						// its subgraph
						if !simDiamond.Contains(newSim) {
							simDiamond.Add(newSim)
							explore(newSim, goroutines)
						}
					}

					if tA.Move == fsa.Recv && tB.Move == fsa.Send && tA.Label == tB.Label {
						// Adds a new edge in the choreography automaton
						newT := fsa.Transition{
							Move:  fsa.Empty,
							Label: fmt.Sprintf("%s \u2192 %s", participantB.Name, participantA.Name),
						}
						fmt.Println(newT)
						choreographyAutomaton.AddTransition(oldSimId, newSimId, newT)

						// If newSim isn't already contained in the simulation diamond then is added and a
						// recursive call on this newly found system configuration is done to explore
						// its subgraph
						if !simDiamond.Contains(newSim) {
							simDiamond.Add(newSim)
							explore(newSim, goroutines)
						}
					}

				})
			})
		}
	}
}
