package main

func main() {
	scanner := NewScannerBuilder().
		WithDB(db).
		WithDirPath(*dirPath).
		WithHashStrategy(new(SHA256HashStrategy)).
		Build()

	err := scanner.Scan()
}
