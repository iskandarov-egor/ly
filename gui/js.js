var canvas = document.getElementById("canvas")
var ctx = canvas.getContext('2d')
var div = document.getElementById('canvasDiv')
let _1vh = window.innerHeight / 100
console.log("HELLOO")
client.onMessage = function(mtype, msg) {
	console.log("GOT MSG", msg);
	MSG = msg;
	switch (mtype) {
		case 44002: // image
			var idata = new ImageData(new Uint8ClampedArray(msg.RGBA), msg.W, msg.H)
			console.log("PUTDATA")
			mainImage.putImageData(idata, msg.X, msg.Y)
			mainView.setImage(mainImage)
			break
		case 44003: // canvas size
			mainImage = new MyImage(msg.W, msg.H)
			mainView.setImage(mainImage)
			break
		case 44005: // draw line
			mainImage.drawLine(
				msg.X1, msg.Y1, msg.X2, msg.Y2, 255*msg.R, 255*msg.G, 255*msg.B)
			mainView.setImage(mainImage)
			break
		default:
			console.error('unk message', mtype)
	}
}

var state = new UIState()
var mainImage = new MyImage(100, 100)
var mainView = new MainView(div, canvas)
mainView.setImage(mainImage)

function change(ya) {
	let tool = ya.value
	state.tool = tool
}

function renderClick() {
	client.sendRenderCommand(state.selection)
}

// set stub image
{
	var canvas = document.createElement('canvas');
	var context = canvas.getContext('2d');
	let img = imgStub;
	img.crossOrigin = "Anonymous";
	canvas.width = img.width;
	canvas.height = img.height;
	context.drawImage(img, 0, 0 );
	var myData = context.getImageData(0, 0, img.width, img.height);
	mainImage = new MyImage(img.width, img.height)
	mainImage.putImageData(myData)
	mainView.setImage(mainImage)
}
