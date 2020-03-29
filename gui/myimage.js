class MyImage {
	constructor(w, h) {
		this.w = w
		this.h = h
		this.canvas = document.createElement('canvas')
		this.canvas.width = w
		this.canvas.height = h
		this.ctx = this.canvas.getContext('2d')
		console.log('CTX', this.ctx)
	}

	putImageData(idata, x, y) {
		this.ctx.putImageData(idata, x, y)
	}

	drawLine(x1, y1, x2, y2, r, g, b) {
		this.ctx.beginPath()
		this.ctx.strokeStyle = `rgb(${r},${g},${b})`
		this.ctx.moveTo(x1, y1)
		this.ctx.lineTo(x2, y2)
		this.ctx.stroke()
		this.ctx.closePath()
		console.log("LINE DRAWN")
	}
}
