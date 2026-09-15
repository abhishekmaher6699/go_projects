package main

import (
	"fmt"
	"sync"
	"time"
)

type Job struct {
	ID      int
	Data    string
	Status  string
	Retries int
}

func worker(
	workerID int,
	jobsChannel <-chan *Job,
	ackChannel chan<- *Job,
	failureChannel chan<- *Job,
	workerFailureChannel chan<- int,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	fmt.Printf("Worker %d waiting for job...\n", workerID)

	for job := range jobsChannel {
		fmt.Printf("Worker %d received Job %d\n", workerID, job.ID)

		job.Status = "processing"

		if job.ID == 5 {
			fmt.Println("💥 A worker crashed!")
			failureChannel <- job
			workerFailureChannel <- workerID
			return
		}

		fmt.Printf("Worker %d STARTED Job %d: %s\n",
			workerID, job.ID, job.Data)

		time.Sleep(2 * time.Second)

		job.Status = "completed"
		ackChannel <- job

		fmt.Printf("Worker %d FINISHED Job %d\n",
			workerID, job.ID)
	}
}

func coordinator(
	jobsChannel chan *Job,
	ackChannel chan *Job,
	failureChannel chan *Job,
	workerFailureChannel chan int,
	jobs []Job,
	wg *sync.WaitGroup,
) {
	for i := range jobs {
		fmt.Println("Sending Job", jobs[i].ID)
		jobsChannel <- &jobs[i]
		fmt.Println("Sent Job", jobs[i].ID)
	}

	completed := 0
	newWorkerID := 3

	for completed < len(jobs) {

		select {
		case job := <-ackChannel:
			fmt.Println("Coordinator received ack", job.ID)
			fmt.Println("Status:", job.Status)
			completed++

		case failedJob := <-failureChannel:

			failedJob.Retries++

			if failedJob.Retries <= 3 {
				fmt.Println("Coordinator received failure for Job", failedJob.ID)
				fmt.Println("Retrying Job", failedJob.ID, "Retry count:", failedJob.Retries)
				jobsChannel <- failedJob
			} else {
				failedJob.Status = "failed"
				fmt.Println("Job", failedJob.ID, "failed after 3 retries. Marking as failed.")
				completed++
			}

		case workerID := <-workerFailureChannel:
			
			fmt.Println("Coordinator detected worker failure. Worker ID:", workerID)

			newWorkerID++

			fmt.Println("Starting replacement worker", newWorkerID)

			startWorker(newWorkerID, jobsChannel, ackChannel, failureChannel, workerFailureChannel, wg)
		}

	}

}

func startWorker(
	workerID int,
	jobsChannel chan *Job,
    ackChannel chan *Job,
    failureChannel chan *Job,
    workerFailureChannel chan int,
    wg *sync.WaitGroup,
) {
	wg.Add(1)

	go worker(
		workerID,
		jobsChannel,
		ackChannel,
		failureChannel,
		workerFailureChannel,
		wg,
	)
}

func main() {

	var wg sync.WaitGroup

	jobs := []Job{
		{ID: 1, Data: "send email", Status: "pending", Retries: 0},
		{ID: 2, Data: "resize image", Status: "pending", Retries: 0},
		{ID: 3, Data: "generate report", Status: "pending", Retries: 0},
		{ID: 4, Data: "process payment", Status: "pending", Retries: 0},
		{ID: 5, Data: "backup database", Status: "pending", Retries: 0},
		{ID: 6, Data: "validate user", Status: "pending", Retries: 0},
		{ID: 7, Data: "upload file", Status: "pending", Retries: 0},
		{ID: 8, Data: "compress video", Status: "pending", Retries: 0},
		{ID: 9, Data: "send notification", Status: "pending", Retries: 0},
		{ID: 10, Data: "generate invoice", Status: "pending", Retries: 0},
		{ID: 11, Data: "sync database", Status: "pending", Retries: 0},
		{ID: 12, Data: "process image", Status: "pending", Retries: 0},
		{ID: 13, Data: "update user profile", Status: "pending", Retries: 0},
		{ID: 14, Data: "clean temporary files", Status: "pending", Retries: 0},
		{ID: 15, Data: "generate analytics", Status: "pending", Retries: 0},
		{ID: 16, Data: "check inventory", Status: "pending", Retries: 0},
		{ID: 17, Data: "export user data", Status: "pending", Retries: 0},
		{ID: 18, Data: "refresh cache", Status: "pending", Retries: 0},
		{ID: 19, Data: "send weekly summary", Status: "pending", Retries: 0},
		{ID: 20, Data: "archive old records", Status: "pending", Retries: 0},
	}

	jobsChannel := make(chan *Job, 20)
	ackChannel := make(chan *Job, 20)
	failureChannel := make(chan *Job, 20)
	workerFailureChannel := make(chan int, 10)

	startWorker(1, jobsChannel, ackChannel, failureChannel, workerFailureChannel, &wg)
	startWorker(2, jobsChannel, ackChannel, failureChannel, workerFailureChannel, &wg)
	startWorker(3, jobsChannel, ackChannel, failureChannel, workerFailureChannel, &wg)

	coordinator(
		jobsChannel,
		ackChannel,
		failureChannel,
		workerFailureChannel,
		jobs,
		&wg,
	)

	close(jobsChannel)

	wg.Wait()

}
