<template>
	<canvas ref="cv" class="cv-image-overlay" :width="width" :height="height"></canvas>
</template>

<script>
import { mapState } from 'vuex'


function align(x) { return ~~(x)+0.5 }

function draw_corner_box(ctx, cx, cy, sx, sy, corner) {
	const left = cx - sx/2
	const right = cx + sx/2
	const top = cy - sy/2
	const bottom = cy + sy/2

	ctx.beginPath()
	ctx.moveTo(align(left + corner), align(top))
	ctx.lineTo(align(left), align(top))
	ctx.lineTo(align(left), align(top+corner))

	ctx.moveTo(align(left + corner), align(bottom))
	ctx.lineTo(align(left), align(bottom))
	ctx.lineTo(align(left), align(bottom-corner))


	ctx.moveTo(align(right-corner), align(bottom))
	ctx.lineTo(align(right), align(bottom))
	ctx.lineTo(align(right), align(bottom-corner))

	ctx.moveTo(align(right-corner), align(top))
	ctx.lineTo(align(right), align(top))
	ctx.lineTo(align(right), align(top+corner))

	ctx.stroke()
}

	
export default {
	props: {
		width: {},
		height: {},
		exif: {},
	},
	data() {
		return {

		}
	},
	computed: {
		canvas() { return this.$refs.cv },
		ctx() { return this.canvas.getContext('2d') },
		rot() {
			if (!('Orientation' in this.exif)) {
				return this.exif.ImageWidth < this.exif.ImageHeight
			}
			return this.exif.Orientation.includes('CW')
		},
		nfaces() {
			if (! ('FacesDetected' in this.exif)) {
				return 0
			}
			return parseInt(this.exif.FacesDetected)
		},
		n_focal_plane() {
			if (! ('FocalPlaneAFPointsUsed' in this.exif)) {
				return 0
			}
			return parseInt(this.exif.FocalPlaneAFPointsUsed)
		},
		canonAF() {
			if (!('AFAreaXPositions' in this.exif)) {
				return null
			}

			const af = {
				AFWidth: this.exif.AFImageWidth,
				AFHeight: this.exif.AFImageHeight,
				x: this.exif.AFAreaXPositions.split(' ').map(n=>parseInt(n)+this.exif.AFImageWidth/2),
				y: this.exif.AFAreaYPositions.split(' ').map(n=>parseInt(n)+this.exif.AFImageHeight/2),
			}

			if ('AFAreaHeights' in this.exif) {
				af.heights = this.exif.AFAreaHeights.split(' ').map(n=>parseInt(n))
			} else if ('AFAreaHeight' in this.exif) {
				af.heights = Array(af.x.length).fill(parseInt(this.exif.AFAreaHeight))
			}

			if ('AFAreaWidths' in this.exif) {
				af.widths = this.exif.AFAreaWidths.split(' ').map(n=>parseInt(n))
			} else if ('AFAreaWidth' in this.exif) {
				af.widths = Array(af.x.length).fill(parseInt(this.exif.AFAreaWidth))
			}


			return af

		},
		...mapState('interface', ['active_layers'])
	},
	methods: {
		draw() {
			this.ctx.clearRect(0,0,this.exif.ImageWidth, this.exif.ImageHeight)

			// focus plane dots (often a grid)
			if (this.active_layers.indexOf('Focus Plane') !== -1) {
				this.draw_focal_plane()
			}

			if (this.active_layers.indexOf('Faces') !== -1 && this.nfaces > 0) {
				this.drawFaces()
			}

			if (this.active_layers.indexOf('AF Intent') !== -1) {
				this.draw_af()
			}

			if (this.active_layers.indexOf('Focus Point') !== -1) {
				this.draw_focus()
			}

		},
		drawFaces() {
			for (let i=0; i<this.nfaces; i++) {
				if ( ! ('Face'+(i+1)+'Position' in this.exif)) {
					continue
				}
				const pts = this.exif['Face'+(i+1)+'Position'].split(' ').map(n=>parseInt(n))
				const face = this.correctFaceCoords(pts)

				if (face !== null) {
					this.ctx.strokeStyle = '#00ff00'
					this.ctx.strokeRect(face[0], face[1], face[2], face[3])
				}				
			}
		},
		draw_focus() {
			this.ctx.strokeStyle = "#ff0000"


			// detect canon method
			if ('AFAreaXPositions' in this.exif && 'AFPointsInFocus' in this.exif) {

				const af = this.canonAF
				const selected = (this.exif.AFPointsInFocus+'').split(',').map(n=>parseInt(n))

				for (let i=0; i<selected.length; i++) {
					const j = selected[i]
					if (this.rot) {
						this.ctx.strokeRect(
							align(af.y[j]/af.AFHeight*this.width),
							align(af.x[j]/af.AFWidth*this.height),
							~~(af.heights[j]/af.AFHeight*this.width),
							~~(af.widths[j]/af.AFWidth*this.height)
						)
					} else {
						this.ctx.strokeRect(
							align(af.x[j]/af.AFWidth * this.width),
							align(af.y[j]/af.AFHeight * this.height),
							~~(af.widths[j]/af.AFWidth*this.width),
							~~(af.heights[j]/af.AFHeight*this.height)
						)
					}
				}

				return
			}


			// try sony method
			if ('FocusLocation' in this.exif) {
				const pts = this.exif.FocusLocation.split(' ').map(n => parseInt(n))

				const size = 10
				const x = this.correct_x(pts[2], pts[3], pts[0], pts[1], this.width)
				const y = this.correct_y(pts[2], pts[3], pts[0], pts[1], this.height)

				this.ctx.strokeRect(align(x-size/2), align(y-size/2), size, size)
			}

		},
		draw_af() { // the spot/zone position before recomposing
			this.ctx.strokeStyle = '#0000ff'

			const mode = 'AFAreaMode' in this.exif ? this.exif.AFAreaMode : ''
			const setting = 'AFAreaModeSetting' in this.exif ? this.exif.AFAreaModeSetting : ''


			if ('AFAreaXPositions' in this.exif && 'AFPointsSelected' in this.exif) {
				// use as canon
				//todo: use Mode here

				const af = this.canonAF
				const selected = (this.exif.AFPointsSelected+'').split(',').map(n=>parseInt(n))

				const pad=4
				for (let i=0; i<selected.length; i++) {
					const j = selected[i]
					if (this.rot) {
						this.ctx.strokeRect(
							align(af.y[j]/af.AFHeight*this.width)-pad,
							align(af.x[j]/af.AFWidth*this.height)-pad,
							~~(af.heights[j]/af.AFHeight*this.width)+pad*2,
							~~(af.widths[j]/af.AFWidth*this.height)+pad*2
						)
					} else {
						this.ctx.strokeRect(
							align(af.x[j]/af.AFWidth * this.width)-pad,
							align(af.y[j]/af.AFHeight * this.height)-pad,
							~~(af.widths[j]/af.AFWidth*this.width)+pad*2,
							~~(af.heights[j]/af.AFHeight*this.height)+pad*2
						)
					}
				}

				return
			}


			if (setting === 'Wide') {
				const safe = 40
				draw_corner_box(this.ctx, this.width/2, this.height/2, this.width-safe, this.height-safe, 50)
				return
			}

			if (mode === 'Center' || setting === 'Center') {
				draw_corner_box(this.ctx, this.width/2, this.height/2, 65, 45, 7)
				return
			}

			if (mode === 'Zone' || setting === 'Zone') {
				const zone = 'AFPointSelected' in this.exif ? this.exif.AFPointSelected : ''

				// zones: Center, Right, Left; Top Right, Top, Top Left; Bottom Right, Bottom, Bottom Left

				const ypad = 20
				const xpad = 20

				let y = this.height/2
				if (zone.includes('Top')) {
					y = this.height/3-ypad
				} else if (zone.includes('Bottom')) {
					y = 2*this.height/3+ypad
				}

				let x = this.width/2
				if (zone.includes('Left')) {
					x = this.width/3-xpad
				} else if (zone.includes('Right')) {
					x = 2*this.width/3+xpad
				}


				draw_corner_box(this.ctx, x, y, this.width/2-xpad, this.height/2-ypad, 20)
				return

			}

			if ('FlexibleSpotPosition' in this.exif) {
				let afsize = 20
				if (mode === 'Expanded Flexible Spot') {
					afsize = 15
				}

				const afpw = 640
				const afph = 480
				const pts = this.exif.FlexibleSpotPosition.split(' ').map(n=>parseInt(n))

				let x = this.correct_x(pts[0], pts[1], afpw, afph, this.width)
				let y = this.correct_y(pts[0], pts[1], afpw, afph, this.height)
				
				this.ctx.strokeRect(align(x-afsize/2),align(y-afsize/2),afsize,afsize)
			}




		},
		correctFaceCoords(pts) {
			const h = this.rot ? this.exif.ImageWidth : this.exif.ImageHeight
			const w = this.rot ? this.exif.ImageHeight : this.exif.ImageWidth

			const x = this.rot ? pts[0] : pts[1]
			const y = this.rot ? pts[1] : pts[0]

			const boxw = this.rot ? pts[2] : pts[3]
			const boxh = this.rot ? pts[3] : pts[2]


			return [
				x/w * this.width,
				( this.rot ? h-y-boxh : y)/h * this.height,
				boxw/w * this.width,
				boxh/h * this.height,
			]
		},
		draw_focal_plane() {
			this.ctx.strokeStyle = '#ffff00'


			if ('AFAreaXPositions' in this.exif) {
				// treat as Canon AF Area grid
				const af = this.canonAF
				for (let i=0; i<af.x.length; i++) {
					if (this.rot) {
						this.ctx.strokeRect(
							align(af.y[i]/af.AFHeight*this.width),
							align(af.x[i]/af.AFWidth*this.height),
							~~(af.heights[i]/af.AFHeight*this.width),
							~~(af.widths[i]/af.AFWidth*this.height)
						)
					} else {
						this.ctx.strokeRect(
							align(af.x[i]/af.AFWidth * this.width),
							align(af.y[i]/af.AFHeight * this.height),
							~~(af.widths[i]/af.AFWidth*this.width),
							~~(af.heights[i]/af.AFHeight*this.height)
						)
					}
				}

				return
			}

			// treat as Sony Focal Plane
			if (! ('FocalPlaneAFPointArea' in this.exif)) {
				return
			}
			const area = this.exif.FocalPlaneAFPointArea.split(' ').map(n=>parseInt(n))

			for (let i=0; i<this.n_focal_plane; i++) {
				if (! ('FocalPlaneAFPointLocation'+(i+1))) {
					continue
				}

				const pts = this.exif['FocalPlaneAFPointLocation'+(i+1)].split(' ').map(n=>parseInt(n))

				let x = this.correct_x(pts[0], pts[1], area[0], area[1], this.width)
				let y = this.correct_y(pts[0], pts[1], area[0], area[1], this.height)
				this.ctx.strokeRect(x,y,10,10)
			}
		},
		correct_x(left, top, lmax, tmax, scale) {
			if (!this.rot) {
				return left/lmax * scale
			}

			// determine direction of rotation
			if (this.exif.Orientation.includes('270')) {
				return top/tmax * scale
			}

			return (tmax-top)/tmax * scale

		},
		correct_y(left, top, lmax, tmax, scale) {
			if (!this.rot) {
				return top/tmax * scale
			}

			if (this.exif.Orientation.includes('270')) {
				return (lmax-left)/lmax * scale
			}
			return left/lmax * scale
		},
	},
	components: {},
	watch: {
		active_layers() {
			this.draw()
		}
	},
	mounted() {
		this.draw()
	},
}
</script>


<style>

.cv-image-overlay {
	position: absolute;
	top: 0;
	left: 0;
	width: 100%;
	height: 100%;
}

</style>