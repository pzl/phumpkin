<template>
	<div class="mapview" ref="view">
		<canvas ref="canvas" :width="width" :height="height"></canvas>
	</div>
</template>


<script>
import { geoMercator, geoPath, merge, mesh } from 'd3-geo'
import { event, select } from 'd3-selection'
import { zoom, zoomIdentity } from 'd3-zoom'
import { tile, tileWrap } from 'd3-tile'

function geoConvert(dms) {
	let m
	let s

	let sep = dms.includes('deg') ? 'deg' : ','
	let split = dms.split(sep)
	const d = parseInt(split[0])

	if (split[1].includes("'")) {
		split = split[1].split("'")
		m = parseInt(split[0])
		s = parseFloat(split[1].replace(/[a-z'"]/gi, ''))
	} else {
		m = parseFloat(split[1].replace(/[a-z'"]/gi, ''))
		s = 0
	}

	const sign = (dms.includes('N') || dms.includes('E')) ? 1 : -1

	return sign * (d + (m/60) + (s/3600))
}

const tileSize = 256 // pixel size of each tile

export default {
	data() {
		const w = 800
		const h = 600
		return {
			width: w,
			height: h,
			t: {
				x: w/2, // initial x translation
				y: h/2, // initial y translation
				k: 1 // initial scale
			},
			photos: [],
			landData: null,
			boundary: null,
		}
	},
	computed: {
		canvas() { return this.$refs.canvas },
		ctx() { return this.canvas.getContext('2d') },
		bg() { return this.$vuetify.theme.dark ? '#333' : '#fff' },
		fg() { return this.$vuetify.theme.dark ? '#4e7372' : '#4e7372' },
		pb() { return this.$vuetify.theme.dark ? 'black' : '#8c8c8c' }, // photo border color
		url() {
			const dpr = window.devicePixelRatio
			return (x,y,z) => `https://stamen-tiles-${"abc"[Math.abs(x+y) % 3]}.a.ssl.fastly.net/${this.tile_type}/${z}/${x}/${y}${dpr > 1 ? '@2x' : ''}.png`
		},
		tile() {
			return tile()
					.size([this.width, this.height])
					.scale(this.projection.scale()*2*Math.PI)
					.translate(this.projection([0,0]))
					.clampX(false)

		},
		tiles() {
			return this.tls.map(([x,y,z], i, {translate: [tx,ty], scale: k}) => {
				const [wx, wy, wz] = tileWrap([x,y,z])
				return {
					url: this.url(wx,wy,wz),
					x: Math.round((x+tx)*k),
					y: Math.round((y+ty)*k),
					height: k,
					width: k,
				}
			})
		},
		scale(){ return this.t && this.t.k !== null ? this.t.k : 1 },
		tx() { return this.t && this.t.x !== null ? this.t.x : 1 },
		ty() { return this.t && this.t.y !== null ? this.t.y : 1 },
		projection() {
			return geoMercator()
				.scale( ((this.width-1)/2/Math.PI) * this.scale)
				.translate([this.tx, this.ty])
		},
		path() {
			return geoPath(this.projection, this.ctx)
		},
		parsed_photos() {
			if (!this.photos) {
				return []
			}
			return this.photos.map(p => {
				p.x = geoConvert(p.meta.loc.lon)
				p.y = geoConvert(p.meta.loc.lat)
				p.position = this.projection([p.x,p.y])
				return p
			})
		},
	},
	methods: {
		zoomed() {
			this.t = event.transform
			this.draw()
		},
		pw(photo){ return photo.thumbs['small'].width/2 },
		ph(photo) { return photo.thumbs['small'].height/2 },
		draw() {
			this.ctx.clearRect(0,0,this.width, this.height)
			this.ctx.save()
			
			if (this.t) {
				this.ctx.translate(this.t.x, this.t.y)
				this.ctx.scale(this.t.k, this.t.k)
			}

			// draw basic map
			this.ctx.beginPath()
			this.path({ type: 'Sphere' })
			this.ctx.strokeStyle = '#ff0000'
			this.ctx.stroke()

			this.ctx.beginPath()
			this.path(this.landData)
			this.ctx.lineWidth = 0.5
			this.ctx.strokeStyle = '#0000ff'
			this.ctx.stroke()
		}
	},
	mounted() {
		this.width = this.$refs.view.clientWidth
		this.height = this.$refs.view.clientHeight
		const z = zoom().scaleExtent([1,1<<16]).on('zoom', this.zoomed)
		select('.mapview canvas').call(z)
			.call(
				z.transform,
				zoomIdentity
					.translate(this.width/2, this.height/2)
			)
		this.$axios.get('/countries-10m.json').then(d => {
			this.landData = merge(d.data, d.data.objects.countries.geometries)
			this.boundary = mesh(d.data, d.data.objects.countries, (a,b) => a !== b)
			this.draw()
		})
		this.$fetch('/api/v1/query/locations')
			.then(data => { this.photos = data.data.photos })
	},
}
</script>


<style>
.mapview {
	height: 90vh;
	position: relative;
}

.map-type-select {
	position: absolute;
	top: 20px;
	left: 20px;
}

.photo-view rect {
	fill: none;
	stroke-width: 4;
}

.photo-spot {
	fill: #ff0000;
}

#attrib {
	position: absolute;
	bottom: 5px;
	right: 5px;
	font: 10px sans-serif;
	padding: 3px;
	opacity: 0.8;
}

#attrib a {
	color: #000;
	font-weight: 700;
	text-decoration: none;
}

</style>