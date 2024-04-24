package main

import (
	"bytes"
	"encoding/hex"
	"sync"
	"time"
	"xatum-reference-miner/config"
	"xatum-reference-miner/log"
	"xatum-reference-miner/xatum"
	"xatum-reference-miner/xatum/client"
	"xatum-reference-miner/xelishash"
	"xatum-reference-miner/xelisutil"
)

// Job is a fast & efficient struct used for storing a job in memory
type Job struct {
	Blob   xelisutil.BlockMiner
	Diff   uint64
	Target [32]byte
}

var cl *client.Client
var connected = false

func main() {

	start = time.Now()

	for i := 0; i < 4; i++ {
		go miningThread()
	}

	go logHashrate()

	clientHandler()
}

func clientHandler() {
	for {
		var err error
		cl, err = client.NewClient("127.0.0.1:6969")
		if err != nil {
			log.Errf("%v", err)
			time.Sleep(time.Second)
			continue
		}

		err = cl.Send(xatum.PacketC2S_Handshake, xatum.C2S_Handshake{
			Addr:  config.ADDRESS,
			Work:  "x",
			Agent: "XelMiner ALPHA",
			Algos: []string{config.ALGO},
		})
		if err != nil {
			log.Err(err)
			time.Sleep(time.Second)
			continue
		}

		go readjobs(cl.Jobs)
		cl.Connect()

		time.Sleep(time.Second)
	}
}

var numHashes float64
var start time.Time
var curJob Job
var mutCurJob sync.RWMutex

func readjobs(clJobs chan xatum.S2C_Job) {
	for {
		job, ok := <-clJobs
		if !ok {
			return
		}

		mutCurJob.Lock()
		connected = true

		curJob = Job{
			Blob:   xelisutil.BlockMiner(job.Blob),
			Diff:   job.Diff,
			Target: xelisutil.GetTargetBytes(job.Diff),
		}
		mutCurJob.Unlock()

		log.Infof("new job: diff %d, blob %x", job.Diff, job.Blob)

	}
}

func miningThread() {
	var scratch = xelishash.ScratchPad{}

	for {
		if !connected {
			time.Sleep(time.Second)
			continue
		}

		if cl == nil {
			log.Err("client is nil")
			time.Sleep(time.Second)
			continue
		}

		mutCurJob.Lock()
		numHashes++
		nonce := curJob.Blob.GetNonce()
		if nonce == 0xffffffffffffffff {
			nonce = 0
		} else {
			nonce++
		}

		curJob.Blob.SetNonce(nonce)
		job := curJob
		mutCurJob.Unlock()

		hash := job.Blob.PowHash(&scratch)

		if bytes.Compare(hash[:], job.Target[:]) < 0 {
			log.Infof("found block with PoW %x", hash)
			cl.Lock()

			cl.Submit(xatum.C2S_Submit{
				Data: job.Blob[:],
				Hash: hex.EncodeToString(hash[:]),
			})

			cl.Unlock()

		}
	}
}

func logHashrate() {
	for {
		mutCurJob.Lock()

		log.Infof("hashrate: %.0f", numHashes/float64(time.Since(start).Milliseconds()/1000))

		numHashes = 0
		start = time.Now()

		mutCurJob.Unlock()

		time.Sleep(10 * time.Second)
	}
}
