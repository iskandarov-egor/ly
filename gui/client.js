var buffer = msgpack.encode({foo: "bar"});

class Client {
	constructor() {
		this.ok = false
		this.timer = undefined
		this.onMessage = function(){}
	}

	connect() {
		if (this.socket) {
			this.socket.onopen = undefined
			this.socket.onclose = undefined
			this.socket.onerror = undefined
			this.socket.onmessage = undefined
			this.socket.close()
		}
		this.socket = new WebSocket('ws://127.0.0.1:8080')
		this.socket.binaryType = 'arraybuffer'
		this.ok = false
		var that = this
		this.socket.addEventListener('open', function (event) {
			console.log("connected to server")
			that.ok = true
		})
		this.socket.addEventListener('close', function (event) {
			//console.log("lost connection to server")
			that.ok = false
		})
		this.socket.addEventListener('error', function(event) {
			//console.error("webSocket error observed:", event)
		})
		this.socket.addEventListener('message', function (msg) {
			console.log("recv")
			var d = new msgpack.Decoder()
			var list = []
			d.on('data', (chunk) => { // i hate you, msgpack-lite api
				list.push(chunk)
			})
			DATA = msg.data
			d.decode(new Uint8Array(msg.data))
			d.end()
			if (list.length != 2) {
				console.error('received wrong item number', list.length)
				return
			}
			var mType = list[0]
			LIST = list
			that.onMessage(list[0], list[1])
		})
		if (!this.timer) {
			this.timer = setInterval(function() {that.check(); }, 500)
		}
	}

	check(){
		if (this.socket.readyState == 3) {
			this.connect()
		}
	}

	sendStringCommand(command) {
		console.log("SEND")
		var that = this
		setTimeout(function(){
			var e = msgpack.Encoder()
			e.encode(44001)
			e.encode({"Hello": "zdrasti"})
			console.log("SEND YOLO")
			that.socket.send(e.buffer.slice(0, e.offset))
		})
	}

	sendRenderCommand(selection) {
		console.log("SEND RENDER")
		var that = this
		setTimeout(function(){
			var e = msgpack.Encoder()
			e.encode(44004)
			if (selection == null) {
				e.encode({})
			} else {
				console.log({                                                                 
                    Area: {                                                                
                        Left: selection[0],                                                
                        Top: selection[1],                                                 
                        Right: selection[2],                                               
                        Bottom: selection[3],                                              
                    }                                                                      
                })
				e.encode({
					Area: {
						Left: selection[0],
						Top: selection[1],
						Right: selection[2],
						Bottom: selection[3],
					}
				})
			}
			that.socket.send(e.buffer.slice(0, e.offset))
		})
	}
}

var LIST = null
var DATA = null
var client = new Client()
client.connect()
var MSG = undefined
