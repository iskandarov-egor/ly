var lblCoords = document.getElementById('lblCoords')
var lblSelection = document.getElementById('lblSelection')
function setCoordsLabel(x, y) {
	function pad(num, size){ return ('            ' + Math.floor(num)).substr(-size); }
	lblCoords.innerHTML = `[${pad(x, 4)}, ${pad(y, 4)}]`
}
function setSelectionLabel(left, top, right, bottom) {
	function pad(num, size){ return ('            ' + Math.floor(num)).substr(-size); }
	lblSelection.innerHTML = `{[${pad(left, 4)}, ${pad(top, 4)}]..[${pad(right, 4)}, ${pad(bottom, 4)}]}`
}
