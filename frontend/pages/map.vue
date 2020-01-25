<template>
	<div class="mapview" ref="view">
		<svg :width="width" :height="height">
			<path class="sphere" :d="sphereData" :transform="t" :fill="bg" />
			<path class="land" :d="land" :transform="t" :fill="fg" />
			<path class="countries boundary" :d="countries" :transform="t" :stroke="bg" />
			<path v-if="t && t.k > 2" class="states boundary" :d="states" :transform="t" :stroke="bg" />
			<path v-if="t && t.k > 8" class="counties boundary" :d="counties" :transform="t" :stroke="bg" />
		</svg>
	</div>
</template>


<script>
import { merge, mesh } from 'topojson-client'
import { geoMercator, geoPath } from 'd3-geo'
import { event, select } from 'd3-selection'
import { zoom } from 'd3-zoom'


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
		}
	},
	computed: {
		bg() { return this.$vuetify.theme.dark ? '#333' : '#fff' },
		fg() { return this.$vuetify.theme.dark ? '#4e7372' : '#4e7372' },
		projection() {
			return geoMercator().translate([this.width/2, this.height/2]).scale((this.width-1)/2/Math.PI)
		},
		path() { return geoPath().projection(this.projection) },
		sphereData() { return this.path({ type: 'Sphere' }) },
		level() {
			if (this.t && this.t.k < 7) {
				return '110'
			} else if (this.t && this.t.k < 30) {
				return '50'
			}
			return '10'
		},
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
		const z = zoom().scaleExtent([1,200]).on('zoom', this.zoomed)
		select('.mapview svg').call(z)
		this.load()
		this.loadUS()
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

</style>