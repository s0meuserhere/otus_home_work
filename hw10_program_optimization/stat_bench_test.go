package hw10programoptimization

import (
	"archive/zip"
	"bytes"
	"io"
	"testing"
)

func BenchmarkGetDomainStat(b *testing.B) {
	r, err := zip.OpenReader("testdata/users.dat.zip")
	if err != nil {
		b.Fatal(err)
	}
	defer r.Close()

	f, err := r.File[0].Open()
	if err != nil {
		b.Fatal(err)
	}

	content, err := io.ReadAll(f)
	f.Close()
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := GetDomainStat(bytes.NewReader(content), "biz")
		if err != nil {
			b.Fatal(err)
		}
	}
}
