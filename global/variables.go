package global

/**
VAR THAT CAN BE CHANGED
*/

// If the video is to be posted to Youtube or not.
var PostVideo bool = false
var ChannelName string = "QuotePixel"

// Flag for Youtube if the video is made for kids or not
var MadeForKids bool = false

// AI voice
var VoiceID string = "Scarlett"
var VoiceSpeed string = "0" // -1 to 1
var VoicePitch string = "1" // 0.5 to 1.5
var Bitrate string = "192k" // 320k, 256k, 192k, ...

// Text on screen
var BorderThickness int = 10

// if the video is to be deleted when done
var DeleteVideoAfterPost bool = false

// max 15 entries
// make sure to test the new thema to make sure there exsist a quote and video for said thema.
var Themas []string = []string{
	"love",
	"friendship",
	"happiness",
	"life",
	"courage",
	"trust",
}

/**
DO NOT CHANGE
*/

// The Youtube # limit
const TagLimit int = 15
