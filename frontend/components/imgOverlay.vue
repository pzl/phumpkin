<template>
	<canvas ref="cv" class="cv-image-overlay" :width="width" :height="height"></canvas>
</template>

<script>

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
		focus_location() { // actually focused thing
			if (! ('FocusLocation' in this.exif)) {
				return null
			}


			const pts = this.exif.FocusLocation.split(' ').map(n => parseInt(n))


			const size = 10
			const x = this.correct_x(pts[2], pts[3], pts[0], pts[1], this.width,  size)
			const y = this.correct_y(pts[2], pts[3], pts[0], pts[1], this.height, size)

			return [
				align(x-size/2),
				align(y-size/2),
				size,
				size,
			]
		},
		rot() {
			if (!('CameraOrientation' in this.exif)) {
				return this.exif.ImageWidth < this.exif.ImageHeight
			}
			return this.exif.CameraOrientation.includes('CW')
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
		}
	},
	methods: {
		draw() {
			this.ctx.clearRect(0,0,this.exif.ImageWidth, this.exif.ImageHeight)

			// focus plane dots (often a grid)
			if (this.n_focal_plane > 0) {
				this.draw_focal_plane()
			}

			if (this.nfaces > 0) {
				this.drawFaces()
			}

			this.draw_af()

			// what ended up being the focal point, after any recomposing?
			const foc = this.focus_location
			if (foc !== null) {
				this.ctx.strokeStyle = "#ff0000"
				this.ctx.strokeRect(foc[0], foc[1], foc[2], foc[3])
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
		draw_af() { // the spot/zone position before recomposing
			this.ctx.strokeStyle = '#0000ff'

			const mode = 'AFAreaMode' in this.exif ? this.exif.AFAreaMode : ''
			const setting = 'AFAreaModeSetting' in this.exif ? this.exif.AFAreaModeSetting : ''


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

				let x = this.correct_x(pts[0], pts[1], afpw, afph, this.width, afsize)
				let y = this.correct_y(pts[0], pts[1], afpw, afph, this.height, afsize)
				
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
			if (! ('FocalPlaneAFPointArea' in this.exif)) {
				return
			}
			const area = this.exif.FocalPlaneAFPointArea.split(' ').map(n=>parseInt(n))

			for (let i=0; i<this.n_focal_plane; i++) {
				if (! ('FocalPlaneAFPointLocation'+(i+1))) {
					continue
				}

				const pts = this.exif['FocalPlaneAFPointLocation'+(i+1)].split(' ').map(n=>parseInt(n))

				let x = this.correct_x(pts[0], pts[1], area[0], area[1], this.width, 10)
				let y = this.correct_y(pts[0], pts[1], area[0], area[1], this.height, 10)

				this.ctx.strokeStyle = '#ffff00'
				this.ctx.strokeRect(x,y,10,10)
			}
		},
		correct_x(left, top, lmax, tmax, scale, size) {
			if (!this.rot) {
				return left/lmax * scale
			}

			// determine direction of rotation
			if (this.exif.Orientation.includes('270')) {
				return top/tmax * scale
			}

			return (tmax-top)/tmax * scale /*- size*/

		},
		correct_y(left, top, lmax, tmax, scale, size) {
			if (!this.rot) {
				return top/tmax * scale
			}

			if (this.exif.Orientation.includes('270')) {
				return (lmax-left)/lmax * scale /*- size*/
			}
			return left/lmax * scale
		},
	},
	components: {},
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