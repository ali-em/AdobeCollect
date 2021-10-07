package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"sort"
	"strconv"
	"strings"
)

type XmlDoc struct {
	XMLName   xml.Name  `xml:"root"`
	Signature string    `xml:"Signature"`
	Version   string    `xml:"Version"`
	Flag      string    `xml:"Flag"`
	DataPos   string    `xml:"DataPos"`
	Messages  []Message `xml:"Message"`
}

type Message struct {
	XMLName xml.Name `xml:"Message"`
	Time    int      `xml:"time,attr"`
	Type    string   `xml:"type,attr"`
	Number  int      `xml:"Number"`
}

type FileTime struct {
	Name      string
	Path      string
	StartTime int
	EndTime   int
	OnlyAudio bool
	Type      string
}

const ShellToUse = "bash"

func Shellout(command string) error {
	cmd := exec.Command(ShellToUse, "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}

func main() {
	args := os.Args
	if len(args) != 2 {
		fmt.Println("Usage: AdobeCollect <FolderAddress>")
		os.Exit(1)
	}

	folderName := args[1]
	files, err := ioutil.ReadDir(folderName)
	if err != nil {
		log.Fatal(err)
	}

	var fileTimes []FileTime
	for _, file := range files {
		name := file.Name()
		filePath := path.Join(folderName, name)
		if !strings.Contains(name, ".xml") || !(strings.Contains(name, "cameraVoip") || strings.Contains(name, "screenshare")) {
			continue
		}

		xmlFile, err := os.Open(filePath)

		if err != nil {
			fmt.Println(err)
			continue
		}

		byteValue, _ := ioutil.ReadAll(xmlFile)

		var xmlDoc XmlDoc

		err = xml.Unmarshal(byteValue, &xmlDoc)

		if err != nil {
			fmt.Println(err)
			continue
		}
		endMessage := xmlDoc.Messages[len(xmlDoc.Messages)-2]

		startTime := endMessage.Number - endMessage.Time
		endTime := endMessage.Number
		onlyAudio := !strings.Contains(xmlDoc.Flag, "video") && strings.Contains(xmlDoc.Flag, "audio")
		var t string
		if strings.Contains(name, "cameraVoip") {
			t = "camera"
		} else {
			t = "screen"
		}
		fileTimes = append(fileTimes,
			FileTime{Name: name[0 : len(name)-4], Path: filePath[0:len(filePath)-4] + ".flv",
				StartTime: startTime, EndTime: endTime, OnlyAudio: onlyAudio, Type: t})
		// defer the closing of our xmlFile so that we can parse it later on
		defer xmlFile.Close()
	}

	sort.Slice(fileTimes, func(i, j int) bool {
		return fileTimes[i].StartTime < fileTimes[j].StartTime
	})

	var cameras []FileTime
	var screens []FileTime

	for _, f := range fileTimes {
		if f.Type == "camera" {
			cameras = append(cameras, f)
		} else {
			screens = append(screens, f)
		}
	}

	var tempCameraVideoPath string
	var cameraVideoStart int
	var cameraVideoLength int
	hasVideo := false

	for i, f := range cameras {
		fmt.Println(i, f.Path, f.OnlyAudio, f.StartTime, f.EndTime)
		if !f.OnlyAudio {
			fmt.Println("Going to convert " + f.Path)
			newName := f.Path + ".mkv"
			tempCameraVideoPath = newName
			cameraVideoStart = f.StartTime
			cameraVideoLength = f.EndTime - f.StartTime
			cmd := "ffmpeg -y -i " + f.Path + " -preset ultrafast " + newName
			fmt.Println(cmd)
			err = Shellout(cmd)
			if err != nil {
				fmt.Println(err)
			}
			hasVideo = true
			break
		}
	}
	if len(tempCameraVideoPath) == 0 {
		f := cameras[0]
		tempCameraVideoPath = f.Path
		cameraVideoStart = f.StartTime
		cameraVideoLength = f.EndTime - f.StartTime
	}

	// Merge Student Audios...
	var cmd string

	cmd = "ffmpeg -y -i " + tempCameraVideoPath

	for i, f := range cameras {
		if !hasVideo && i == 0 {
			continue
		}
		if f.OnlyAudio {
			cmd += " -i " + f.Path
		}
	}
	cmd += " -filter_complex \""
	for i, f := range cameras {

		if f.OnlyAudio {
			cmd +=
				"[" + strconv.Itoa(i) + ":a]adelay=" + strconv.Itoa(f.StartTime) +
					"|" + strconv.Itoa(f.StartTime) + "[" + f.Name + "];"
		}
	}

	if hasVideo {
		cmd += " [0:a]"
	}
	for _, f := range cameras {
		if f.OnlyAudio {
			cmd += "[" + f.Name + "]"
		}
	}
	joinedAudiosFile := path.Join(folderName, "Camera.mkv")
	cmd += "amix=" + strconv.Itoa(len(cameras)) + "[a]\" "
	if hasVideo {
		cmd += " -map 0:v "
	}
	cmd += "-map \"[a]\"  -c:v copy -preset ultrafast " + joinedAudiosFile

	fmt.Println(cmd)
	err = Shellout(cmd)
	if err != nil {
		fmt.Println(err)
	} else {
		if hasVideo {
			os.Remove(tempCameraVideoPath)
		}
	}

	fmt.Println("Screen sharings: ", len(screens))
	if len(screens) == 0 {
		fmt.Println("Finished successfully, No Screen sharing")
		os.Rename(joinedAudiosFile, path.Join(folderName, "Final.mkv"))
		return
	}

	cmd = "ffmpeg -y -err_detect ignore_err -i " + joinedAudiosFile
	for _, s := range screens {
		cmd += " -i " + s.Path
	}
	cmd += " -filter_complex \"color=s=1280x720:c=black [base]; " // Black base
	cmd += " [0:v] setpts=PTS-STARTPTS+" + strconv.Itoa(cameraVideoStart) +
		", scale=280x280 [lowerright];" // Camera

	// Screens
	for i, s := range screens {

		cmd += "[" + strconv.Itoa(i+1) + ":" + "v] setpts=PTS-STARTPTS+" + strconv.Itoa(s.StartTime) +
			", scale=1280x720 [" + s.Name + "];"
	}

	cmd += "[base]"
	for i, s := range screens {
		if i > 0 {
			cmd += "[tmp" + strconv.Itoa(i-1) + "]"
		}
		cmd += "[" + s.Name + "] overlay"
		cmd += "=[tmp" + strconv.Itoa(i) + "];"
	}
	cmd += "[tmp" + strconv.Itoa(len(screens)-1) + "]" + "[lowerright] overlay=x=1000:y=480\""

	cmd += " -t " + strconv.Itoa(cameraVideoLength/1000) + " -c:a mp2 -preset ultrafast " + path.Join(folderName, "Final.mkv")
	fmt.Println("\n\n", cmd)
	err = Shellout(cmd)
	if err != nil {
		fmt.Println(err)
	}
}
