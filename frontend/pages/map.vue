<template>
	<div class="mapview">
		<svg :width="width" :height="height">
			<path class="sphere" :d="sphereData" :transform="t" />
			<path class="land" :d="landData" :transform="t" />
			<path class="boundary" :d="boundary" :transform="t" />
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
			landData: null,
			boundary: null,
			t: null,
		}
	},
	computed: {
		projection() {
			return geoMercator().translate([this.width/2, this.height/2]).scale((this.width-1)/2/Math.PI)
		},
		path() {
			return geoPath().projection(this.projection)
		},
		sphereData() {
			return this.path({ type: 'Sphere' })
		}
	},
	methods: {
		zoomed() {
			this.t = event.transform
		},
	},
	mounted() {
		const z = zoom().scaleExtent([1,200]).on('zoom', this.zoomed)
		select('.mapview svg').call(z)

		this.$axios.get('/countries-10m.json').then(d => {
			this.landData = this.path(merge(d.data, d.data.objects.countries.geometries))
			this.boundary = this.path(mesh(d.data, d.data.objects.countries, (a,b) => a !== b))
		})

	},
}
</script>


<style>
.sphere {
	fill: #fff;
}

.land {
	fill: #4e7372;
}

.boundary {
	fill: none;
	stroke: #fff;
	stroke-linejoin: round;
	stroke-linecap: round;
	vector-effect: non-scaling-stroke;
}

</style>