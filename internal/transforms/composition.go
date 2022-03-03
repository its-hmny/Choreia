// Copyright 2020 Enea Guidi (hmny). All rights reserved.
// This file are distributed under the General Public License v 3.0.
// A copy of abovesaid license can be found in the LICENSE file.

// Package transforms declares the types and functions used to transform and work with some type of FSA.
// Come of the transformation implemented here are standard such as determinization (Subset Construction),
// minimization but more are specifically related to Choreia (GoroutineFSA extraction & Composition)
//
package transforms

import (
	"log"

	list "github.com/emirpasic/gods/lists/singlylinkedlist"
	"github.com/its-hmny/Choreia/internal/data_structures/fsa"
)

type tmp struct {
	GoroutineFSA
	currentState int
}

func ComposeGoroutines(goroutines map[string]GoroutineFSA) *fsa.FSA {
	mainGrFSA, exist := goroutines["main (0)"]

	if !exist {
		log.Fatal("Could not find GoroutineFSA for 'main'")
	}

	entrypoint := list.New(tmp{mainGrFSA, 0})
	choreographyAutomaton := fsa.New()
	explore(entrypoint, choreographyAutomaton)
	return choreographyAutomaton
}

func explore(sim *list.List, automaton *fsa.FSA) {

	for indexA, itemA := range sim.Values() {
		participantA := itemA.(tmp)

		// Unary transitions handling (Spawn)
		participantA.Automaton.ForEachTransition(func(from, to int, t fsa.Transition) {
			// Only interested in unary transitions from the current state
			if from != participantA.currentState || t.Move != fsa.Spawn {
				return
			}

			// Makes a copy of the current participant adn updates its current state
			copy := participantA
			copy.currentState = to
			// TODO Creates a new ... with the new participant state instead of the old one
			newSim := list.New(sim.Values())
			newSim.Remove(indexA)
			newSim.Insert(indexA, participantA)

			// ? collega newSim a sim nell'FSA
			// TODO Add edges and states in automaton

			// ? se newSim non esiste già allora aggiungilo alla map e recursive call
			// TODO Recursive call on new ... explore(newSim, automaton)
		})

		// Binary transitions handling (Send, Recv)
		for indexB, itemB := range sim.Values() {
			participantB := itemB.(tmp)

			participantA.Automaton.ForEachTransition(func(fromA, toA int, tA fsa.Transition) {
				participantB.Automaton.ForEachTransition(func(fromB, toB int, tB fsa.Transition) {
					// Makes a copy of both A and B and updates their respective "currentstate" fields
					copyA, copyB := participantA, participantB
					copyA.currentState, copyB.currentState = toA, toB

					// TODO Creates a new ... with the new participant state instead of the old one
					newSim := list.New(sim.Values())
					newSim.Remove(indexA)
					newSim.Remove(indexB)
					newSim.Insert(indexA, copyA)
					newSim.Insert(indexB, copyB)

					if tA.Move == fsa.Send && tB.Move == fsa.Recv && tA.Label == tB.Label {
						// ? collega newSim a sim nell'FSA
						// TODO Add edges and states in automaton

					}

					if tA.Move == fsa.Recv && tB.Move == fsa.Send && tA.Label == tB.Label {
						// ? collega newSim a sim nell'FSA
						// TODO Add edges and states in automaton

					}

					// ? se newSim non esiste già allora aggiungilo alla map e recursive call
					// TODO Recursive call on new ... explore(newSim, automaton)
				})
			})
		}
	}
}
