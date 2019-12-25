<template>
	<v-card outlined  class="ma-2 mb-5 summary-card" :max-height="600">
		<v-card-title>{{name}} <sup><v-icon small v-if="meta.exif.Make" :title="meta.exif.Model">mdi-{{ camera_icon }}</v-icon></sup></v-card-title>
		<v-card-subtitle><rating :readonly="true" :value="meta.rating" /></v-card-subtitle>

		<v-card-text>
			<div>{{ sizeof }}</div>

			<v-row>
				<v-icon v-if="batt" :title="meta.exif.BatteryLevel" small>{{batt }}</v-icon>
				<v-icon v-if="metering" :title="'Metering: '+meta.exif.MeteringMode" small>{{metering}}</v-icon>
				<v-icon v-if="flash" :title="meta.exif.Flash" small>{{ flash }}</v-icon>
				<v-icon v-if="temp" :title="temp" small>mdi-thermometer</v-icon>
				<v-icon v-if="meta.exif.FacesDetected" :title="meta.exif.FacesDetected" small>mdi-face-recognition</v-icon>
			</v-row>

			<div v-if="meta.exif">
				<v-icon small>mdi-camera-iris</v-icon> <v-icon v-if="meta.exif.SelfTimer && meta.exif.SelfTimer !== 'Off'" small>mdi-camera-timer</v-icon>
				<div>f / {{ meta.exif.Aperture }}</div>
				<div>{{ exposure }}</div>
				<div>ISO: {{ meta.exif.ISO }}</div>
				<div>Focal Length: {{meta.exif.FocalLength}}</div>
				<div>Focus Mode: {{meta.exif.FocusMode}}</div>
				<div>IS: {{meta.exif.ImageStabilization}}</div>
			</div>

			<div v-if="meta.exif.LensID">{{meta.exif.LensID}}</div>

			<div class="loc" v-if="loc"><v-icon small>mdi-map-marker</v-icon> {{ loc.lat }}, {{ loc.lon }}</div>
			<tags v-if="tags.length" :dark="false" :tags="tags" />

			<div v-if="meta.creator">Creator: {{ meta.creator }}</div>
			<div v-if="meta.rights">Rights: {{ meta.rights }}</div>

			<div v-if="meta.history && meta.history.length">
				<div class="d-flex">
					<span>History</span>
					<v-spacer />
					<v-btn icon x-small @click="show_history = !show_history"><v-icon small>mdi-menu-{{show_history ? 'up' :'down'}}</v-icon></v-btn>
				</div>

				<v-expand-transition>
					<div v-if="show_history">
						<v-list dense>
							<v-list-item v-for="(h,i) in meta.history" :key="i" class="pa-0">
									<v-list-item-title style="flex-grow: 11">{{ h.name }}</v-list-item-title>
									<v-list-item-subtitle v-if="h.multi_name">{{ h.multi_name }}</v-list-item-subtitle>
									<v-list-item-icon>
										<v-icon x-small dense color="grey lighten-1">mdi-radiobox-{{h.enabled ? 'marked' : 'blank' }}</v-icon>
									</v-list-item-icon>
							</v-list-item>
						</v-list>
					</div>
				</v-expand-transition>
			</div>
		</v-card-text>

		<v-card-actions>
			<v-spacer />
			<v-menu v-model="show_meta" :close-on-content-click="false" offset-x>
				<template v-slot:activator="{ on }">
					<v-btn v-on="on" icon>
						<v-icon small>mdi-details</v-icon>
					</v-btn>
				</template>


				<v-card :max-height="500">
					<v-card-text class="text--primary xmp-popout">
						<pre>{{ JSON.stringify(meta, null, 2) }}</pre>
					</v-card-text>
				</v-card>
			</v-menu>
		</v-card-actions>

	</v-card>
</template>

<script>
import Rating from '~/components/rating'

export default {
	props: {
		name: {},
		dir: {},
		size: {}, // in bytes
		rating: {},
		tags: {}, // array of strings
		meta: {}, //
		loc: {}, // null or {lat:'', lon:''}
		thumbs: {}, // full: { url: "...", width: n, height: n}
		original: {}, //{ url: "...", width: n, height: n}
	},
	data() {
		return {
			show_meta: false,
			show_history: false,
		}
	},
	computed: {
		sizeof() {
			const units = ["b","KB","MB","GB","TB"]
			let unit = 0
			let b = this.size

			while (b > 1024) {
				b = b/1024
				unit++
			}
			return b.toFixed(2)+" "+units[unit]
		},
		camera_icon() {
			if (!this.meta || !this.meta.exif || !this.meta.exif.Make) {
				return ''
			}
			switch (this.meta.exif.Make) {
				case 'SONY': return 'alpha'
				case 'CANON': return 'alpha-c'
				case 'iPhone': return 'apple'
			}
		},
		batt() {
			if (!this.meta || !this.meta.exif || !this.meta.exif.BatteryLevel) {
				return null
			}
			const lvl = parseInt(this.meta.exif.BatteryLevel.replace('%',''), 10)
			return 'mdi-battery-'+Math.ceil(lvl / 10) * 10
		},
		metering() {
			if (!this.meta || !this.meta.exif || !this.meta.exif.MeteringMode) {
				return null
			}
			switch (this.meta.exif.MeteringMode) {
				case 'Multi-segment': return 'mdi-camera-metering-matrix'
			}
			return null
		},
		flash() {
			if (!this.meta || !this.meta.exif || !this.meta.exif.Flash) {
				return null
			}
			if (this.meta.exif.Flash.startsWith("Off")) {
				return 'mdi-flash-off'
			}
			return null
		},
		exposure() {
			if (!this.meta || !this.meta.exif || !this.meta.exif.ExposureTime) {
				return null
			}
			if ( !(""+this.meta.exif.ExposureTime).includes('/') ) {
				return this.meta.exif.ExposureTime + "s"
			}
			return this.meta.exif.ExposureTime
		},
		temp() {
			if (!this.meta || !this.meta.exif) {
				return null
			}
			return this.meta.exif.AmbientTemperature ||
					this.meta.exif.CameraTemperature ||
					this.meta.exif.BatteryTemperature ||
					null
		},
	},
	components: { Rating }
}
</script>


<style>
.xmp-popout {
	background-color: white;
}

.theme--dark .xmp-popout {
	background-color: #303030;
}

.summary-card {
	overflow-y: auto;
}
</style>