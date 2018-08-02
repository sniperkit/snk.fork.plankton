package main

import (
	"log"

	"github.com/johnshiver/plankton/scheduler"
)

func main() {
	// csv_runner := NewCSVRunner("/Users/js231813/go/src/github.com/johnshiver/plankton/example/postgres_example/the-movies-dataset/movies_metadata.csv")
	// csv_runner2 := NewCSVRunner("/Users/js231813/go/src/github.com/johnshiver/plankton/example/postgres_example/the-movies-dataset/links.csv")
	// csv_runner3 := NewCSVRunner("/Users/js231813/go/src/github.com/johnshiver/plankton/example/postgres_example/the-movies-dataset/credits.csv")
	csv_runner4 := NewCSVRunner("/Users/js231813/go/src/github.com/johnshiver/plankton/example/postgres_example/the-movies-dataset/ratings.csv")

	pg_loader := NewPostgresLoader("some_table")
	pg_loader.GetTask().AddChildren(csv_runner4)

	my_scheduler, err := scheduler.NewTaskScheduler(pg_loader)
	if err != nil {
		log.Panic(err)
	}
	my_scheduler.Start()

}