package main

import (
	"testing"
)

func BenchmarkBufferedChannel1(b *testing.B) {
	size := 1
	for i := 0; i < b.N; i++ {
		bufferedChannel(size)
	}
}

func BenchmarkSoftwareChannel1(b *testing.B) {
	size := 1
	for i := 0; i < b.N; i++ {
		softwareChannel(size)
	}
}

func BenchmarkBufferedChannel2(b *testing.B) {
	size := 2
	for i := 0; i < b.N; i++ {
		bufferedChannel(size)
	}
}

func BenchmarkSoftwareChannel2(b *testing.B) {
	size := 2
	for i := 0; i < b.N; i++ {
		softwareChannel(size)
	}
}

func BenchmarkBufferedChannel200(b *testing.B) {
	size := 200
	for i := 0; i < b.N; i++ {
		bufferedChannel(size)
	}
}

func BenchmarkSoftwareChannel200(b *testing.B) {
	size := 200
	for i := 0; i < b.N; i++ {
		softwareChannel(size)
	}
}

func BenchmarkBufferedChannel1000(b *testing.B) {
	size := 1000
	for i := 0; i < b.N; i++ {
		bufferedChannel(size)
	}
}

func BenchmarkSoftwareChannel1000(b *testing.B) {
	size := 1000
	for i := 0; i < b.N; i++ {
		softwareChannel(size)
	}
}

func BenchmarkBufferedChannel2000(b *testing.B) {
	size := 2000
	for i := 0; i < b.N; i++ {
		bufferedChannel(size)
	}
}

func BenchmarkSoftwareChannel2000(b *testing.B) {
	size := 2000
	for i := 0; i < b.N; i++ {
		softwareChannel(size)
	}
}
