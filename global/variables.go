package global

/**
VAR THAT CAN BE CHANGED
*/

// If the video is to be posted to Youtube or not.
var PostVideo bool = false

// Flag for Youtube if the video is made for kids or not
var MadeForKids bool = false

// The speed of the AI voice
var VoiceSpeed float64 = 1
var Pitch float64 = 5 // range -20 to 20

// if the video is to be deleted when done
var DeleteVideoAfterPost bool = false
var DeleteVideoParts bool = true

var BorderThickness = 5

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
