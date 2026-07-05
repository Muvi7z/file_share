package main

import (
	"context"
	"file_share/internal/service/scan"
)

// TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>
func main() {

	videoService := scan.New(nil, nil)

	err := videoService.ScanFolder(context.Background(), "C:\\Users\\Ochir\\Videos")
	if err != nil {
		return
	}

	//inputPath := "./hls_files/15516447410713.mp4"
	//outpuPath := "./hls_files/out"
	//err := videoService.Segment(inputPath, outpuPath)
	//if err != nil {
	//	log.Fatal(err)
	//	return
	//}
}
