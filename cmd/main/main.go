package main

import (
	"file_share/internal/service/videos"
	"log"
)

// TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>
func main() {

	//videoRepository := repository.NewRepository()

	videoService := videos.NewService(nil)

	inputPath := "./hls_files/15516447410713.mp4"
	outpuPath := "./hls_files/out"
	err := videoService.Segment(inputPath, outpuPath)
	if err != nil {
		log.Fatal(err)
		return
	}
}
