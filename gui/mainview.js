class MainView {
	constructor(div, canvas) {
		this.MX = 0 // mousedown origin xy
		this.MY = 0
		this.TX = 0 // current translate xy
		this.TY = 0
		this.OTX = 0 // original translate xy on mousedown
		this.OTY = 0
		let _1vh = window.innerHeight / 100
		this.S = 100 * _1vh / canvas.height // current scale
		this.div = div
		this.canvas = canvas
		let that = this
		canvas.onmousedown = function(e) { that.onmousedown(e); }
		canvas.onmousemove = function(e) { that.onmousemove(e); }
		canvas.onmousewheel = function(e) { that.onmousewheel(e); }
		this.ctx = canvas.getContext('2d')
		this.image = null
		this.updateTransform()
	}

	updateTransform() {
		var s = 'translate(' + this.TX + 'px,' + this.TY + 'px) translate(-50%, -50%) scale(' + this.S + ') translate(50%, 50%)'
		this.canvas.style.transform = s
		if (state.selection != null) {
			this.ctx.putImageData(this.image.ctx.getImageData(0, 0, this.image.w, this.image.h), 0, 0)
			this.ctx.strokeStyle = 'white'
			this.ctx.lineWidth = 0.5
			let left = state.selection[0]
			let top = state.selection[1]
			let right = state.selection[2]
			let bottom = state.selection[3]
			this.ctx.strokeRect(left - 1, top - 1, right - left + 2, bottom - top + 2)
			setSelectionLabel(left, top, right, bottom)
		}
	}

	setImage(img) {
		this.image = img
		this.canvas.width = img.canvas.width
		this.canvas.height = img.canvas.height
		this.ctx.putImageData(img.ctx.getImageData(0, 0, img.w, img.h), 0, 0)
		this.S = 100 * _1vh / canvas.height
		this.updateTransform()
		console.log('IMAGE SET')
	}

	onmousedown(e) {
		this.MX = e.layerX
		this.MY = e.layerY
		this.OTX = this.TX
		this.OTY = this.TY
		console.log(e)
	}

	view2image(x, y) {
		x = (x- this.TX) / this.S
		y = (y - this.TY) / this.S
		return [x, y]
	}

	onmousemove(e) {
		if (state.tool == 'move') {
			if (e.buttons == 1) {
				this.TX = this.OTX + e.layerX - this.MX
				this.TY = this.OTY + e.layerY - this.MY
				if (this.TX > 0) {
					this.TX = 0
				}
				if (this.TY > 0) {
					this.TY = 0
				}
				this.updateTransform()
			}
			var x = (e.layerX - this.TX) / this.S
			var y = (e.layerY - this.TY) / this.S
			setCoordsLabel(x, y)
		} else if (state.tool == 'select') {
			if (e.buttons == 1) {
				let [sx, sy] = this.view2image(this.MX, this.MY)
				let [ex, ey] = this.view2image(e.layerX, e.layerY)
				if (sx > ex) {
					[sx, ex] = [ex, sx]
				}
				if (sy > ey) {
					[sy, ey] = [ey, sy]
				}
				sx = Math.floor(sx)
				sy = Math.floor(sy)
				ex = Math.floor(ex)
				ey = Math.floor(ey)
				state.selection = [sx, sy, ex, ey]
				this.updateTransform()
			}
		}
	}
	onmousewheel(e){
		if (state.tool == 'move') {
			if (this.canvas.height*this.S > this.div.clientHeight) {
				var x = (e.layerX - this.TX) / this.S
				var y = (e.layerY - this.TY) / this.S
				if (e.deltaY < 0) {
					this.S *= 1.05
				} else {
					this.S /= 1.05
				}
				if (this.S != 1) {
					this.TX = e.layerX - x*this.S
					this.TY = e.layerY - y*this.S
				}
				if (this.TX > 0) {
					this.TX = 0
				}
				if (this.TY > 0) {
					this.TY = 0
				}
			} else {
				if (e.deltaY < 0) {
					this.S *= 1.05
				} else {
					this.S /= 1.05
				}
			}
			this.updateTransform()
		}
	}
}
