<template>
	<div class="mapview" ref="view">
		<svg :width="width" :height="height">
			<image v-for="(t,i) in tiles" :key="t.url+i"
				:href="t.url"
				:x="t.x" :y="t.y"
				:width="t.width" :height="t.height"
			/>
			<circle class="photo-spot" v-for="(p,i) in parsed_photos" :key="'photo'+i" :cx="p.position[0]" :cy="p.position[1]" :r="4">
				<title>{{ p.name }}</title>
			</circle>
		</svg>
		<v-select v-model="tile_type" :items="tile_flavors" dense hide-details return-object class="map-type-select" />
	</div>
</template>


<script>
import { geoMercator, geoPath } from 'd3-geo'
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
		return {
			width: 800,
			height: 600,
			t: null,
			photos: [],
			tile_flavors: [
				'toner',
				'toner-hybrid',
				'toner-labels',
				'toner-lines',
				'toner-background',
				'toner-lite',
				'terrain',
				'terrain-labels',
				'terrain-lines',
				'terrain-background',
				'watercolor'
			],
			tile_type: 'toner-lite',
			tls: [],
		}
	},
	computed: {
		bg() { return this.$vuetify.theme.dark ? '#333' : '#fff' },
		fg() { return this.$vuetify.theme.dark ? '#4e7372' : '#4e7372' },
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
		path() { return geoPath().projection(this.projection) },
		parsed_photos() {
			if (!this.photos) {
				return null
			}
			return this.photos.map(p => {
				p.x = geoConvert(p.meta.loc.lon)
				p.y = geoConvert(p.meta.loc.lat)
				p.position = this.projection([p.x,p.y])
				return p
			})
		}
	},
	methods: {
		zoomed() {
			this.t = event.transform
			this.tls = this.tile(event.transform)
		},
	},
	mounted() {
		this.width = this.$refs.view.clientWidth
		this.height = this.$refs.view.clientHeight
		const z = zoom().scaleExtent([1,1<<16]).on('zoom', this.zoomed)
		select('.mapview svg')
			.call(z)
			.call(
				z.transform,
				zoomIdentity
					.translate(this.width/2, this.height/2)
			)

		this.tls = this.tile()

		let server = location.origin
		if (server === "http://localhost:3000") {
			server = "http://localhost:6001"
		}
		this.$axios.get(server + '/api/v1/locations')
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

.photo-spot {
	fill: #ff0000;
	vector-effect: non-scaling-stroke;
}

</style>