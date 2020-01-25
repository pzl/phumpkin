<template>
	<div class="mapview" ref="view">
		<svg :width="width" :height="height">
			<path class="sphere" :d="sphereData" :transform="t" :fill="bg" />
			<path class="land" :d="land" :transform="t" :fill="fg" />
			<path class="countries boundary" :d="countries" :transform="t" :stroke="bg" />
			<path v-if="t && scale > 2" class="states boundary" :d="states" :transform="t" :stroke="bg" />
			<path v-if="t && scale > 8" class="counties boundary" :d="counties" :transform="t" :stroke="bg" />
			<circle class="photo-spot" v-for="(p,i) in parsed_photos" :key="'photo'+i" :cx="p.position[0]" :cy="p.position[1]" :r="4/scale" :transform="t">
				<title>{{ p.name }}</title>
			</circle>
		</svg>
	</div>
</template>


<script>
import { merge, mesh } from 'topojson-client'
import { geoMercator, geoPath } from 'd3-geo'
import { event, select } from 'd3-selection'
import { zoom } from 'd3-zoom'

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

export default {
	data() {
		return {
			width: 800,
			height: 600,
			land: null,
			countries: null,
			states: null,
			counties: null,
			t: null,
			cache: {},
			photos: [],
		}
	},
	computed: {
		bg() { return this.$vuetify.theme.dark ? '#333' : '#fff' },
		fg() { return this.$vuetify.theme.dark ? '#4e7372' : '#4e7372' },
		scale(){ return this.t && this.t.k !== null ? this.t.k : 1 },
		projection() {
			return geoMercator().translate([this.width/2, this.height/2]).scale((this.width-1)/2/Math.PI)
		},
		path() { return geoPath().projection(this.projection) },
		sphereData() { return this.path({ type: 'Sphere' }) },
		level() {
			if (this.scale < 7) {
				return '110'
			} else if (this.scale < 30) {
				return '50'
			}
			return '10'
		},
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
		},
		load() {
			if (this.level in this.cache) {
				this.land = this.cache[this.level].land
				this.countries = this.cache[this.level].countries
				return
			}
			this.$axios.get(location.origin + '/countries-'+this.level+'m.json').then(d => {
				this.cache[this.level] = {
					land: this.path(merge(d.data, d.data.objects.countries.geometries)),
					countries: this.path(mesh(d.data, d.data.objects.countries, (a,b) => a !== b)),
				}
				this.land = this.cache[this.level].land
				this.countries = this.cache[this.level].countries
			})
		},
		loadUS() {
			this.$axios.get(location.origin+'/states-10m.json').then(d => {
				this.states = this.path(mesh(d.data, d.data.objects.states))
			})
			this.$axios.get(location.origin+'/counties-10m.json').then(d => {
				this.counties = this.path(mesh(d.data, d.data.objects.counties, (a,b) => a !== b))
			})
		}
	},
	watch: {
		level() {
			this.load()
		}
	},
	mounted() {
		this.width = this.$refs.view.clientWidth
		this.height = this.$refs.view.clientHeight
		const z = zoom().scaleExtent([1,600]).on('zoom', this.zoomed)
		select('.mapview svg').call(z)
		this.load()
		this.loadUS()

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
}

.boundary {
	fill: none;
	stroke-linejoin: round;
	stroke-linecap: round;
	vector-effect: non-scaling-stroke;
}

.photo-spot {
	fill: #ff0000;
	vector-effect: non-scaling-stroke;
}

</style>