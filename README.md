# AdobeConnect To .mkv

## Introduction

This app helps downloading AdobeConnect videos in situations like when you can just `watch` the past videos and you can't download it, e.g: University Of Tehran!

## How to use

1. Download AdobeConnect zip output using method in this [link](https://stackoverflow.com/questions/5179517/how-can-i-export-an-adobe-connect-recording-as-a-video):

   - Log into your Adobe Connect account
   - Click on Meetings > My Meetings
   - Click on the link for the recording
   - Click the “Recordings” link (right-side of screen)
   - Click the link in the “Name” column
   - Copy the “URL for Viewing” – Example, http://mycompany.adobeconnect.com/ p12345678/
   - Paste it into a new browser tab then add the following to the end of the URL: output/filename.zip?download=zip
   - Your URL should look similar to this example, http://mycompany.adobeconnect.com/p12345678/output/filename.zip?download=zip
   - Unzip the downloaded zip file

   This will give you a folder full of `.flv` and `.xml` files.

2. Install `ffmpeg`:

   Ubuntu:

   ```
   apt install ffmpeg
   ```

3. [Download](https://github.com/ali-em/AdobeCollect/releases/download/v0.8/AdobeCollect) and run app:

   ```
   ./AdobeCollect <UnzippedfolderPath>
   ```

   Example:
   We assume you have unzipped the downloaded file to `database_session_1` folder and so `database_session_1` folder contains some `flv` and `xml` files:

   ```
   ./AdobeCollect database_session_1
   ```

   As soon as you run this command AdobeCollect will start to convert videos using ffmpeg and output will be in `database_session_1/Final.mkv` after some time.

## How it works

In the default format of downloaded zip, there is 1 flv for each part, for example: one flv for webcam of teacher, one for shared screens, one for when each microphone opens, etc.

The downloaded zip file contains some XML files that includes metadata of flv files, like when each file should start, how long it is, is it just audio or it has video, .... (And even the presenter exact OS version!!!)

We use `ffmpeg` to merge all of them together and create a single mkv file.

## Performance

On my own tests it converted an hour of video + presentation + bunch of questioning(microphone openings) in 14 minutes using 2 cores of an old potato i5 cpu :potato: !

I should work on it's performance more but now it's good and so much better than recording the screen!

<i>I hope you enjoy it, Give it :star: if you like it </i>
