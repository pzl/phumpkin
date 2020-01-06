<template>
	<v-card outlined  class="ma-2 mb-5 summary-card" :max-height="600">
		<v-card-title>{{name}} <sup><v-icon small v-if="meta.exif.Make" :title="meta.exif.Model">mdi-{{ camera_icon }}</v-icon></sup></v-card-title>
		<v-card-subtitle><rating :readonly="true" :value="meta.rating" /></v-card-subtitle>

		<v-tabs grow v-model="tab" height="20px">
			<v-tab style="min-width: 20px" class="pa-0" href="#basic"><v-icon small>mdi-calendar-blank</v-icon></v-tab>
			<v-tab style="min-width: 20px" class="pa-0" href="#shot-info"><v-icon small>mdi-camera-iris</v-icon></v-tab>
			<v-tab style="min-width: 20px" class="pa-0" href="#tags"><v-icon small>mdi-tag-heart</v-icon></v-tab>
			<v-tab style="min-width: 20px" class="pa-0" href="#edits" v-if="meta.history && meta.history.length"><v-icon small>mdi-image-edit</v-icon></v-tab>
			<v-tab style="min-width: 20px" class="pa-0" href="#copy"><v-icon small>mdi-copyright</v-icon></v-tab>
		</v-tabs>

		<v-card-text>
			<v-tabs-items v-model="tab">
				<v-tab-item value="basic">
					<div>{{ sizeof }}</div>
					<div>{{ original.width }}x{{ original.height }}</div>
					<div v-if="taken">{{taken}}</div>
					<div class="loc" v-if="meta.loc"><v-icon small>mdi-map-marker</v-icon> {{ meta.loc.lat }}, {{ meta.loc.lon }}<span v-if="meta.loc.alt">, {{ meta.loc.alt }}</span></div>
				</v-tab-item>
				<v-tab-item value="shot-info">
					<v-row no-gutters justify="space-between">
						<v-col cols="1" v-if="metering">
							<v-icon :title="'Metering: '+meta.exif.MeteringMode" small>{{ metering }}</v-icon>
						</v-col>
						<v-col cols="1" v-if="focus_icon">
							<v-icon :title="meta.exif.AFAreaMode" small>{{ focus_icon }}</v-icon>
						</v-col>
						<v-col cols="1" v-if="flash">
							<v-icon :title="meta.exif.Flash" small>{{ flash }}</v-icon>
						</v-col>
						<v-col cols="1" v-if="meta.exif.ImageStabilization">
							<v-icon :title="'IS: '+meta.exif.ImageStabilization" small>mdi-vibrate{{ meta.exif.ImageStabilization.includes("On") ? '' : '-off' }}</v-icon>
						</v-col>
						<v-col cols="1" v-if="temp">
							<v-icon :title="temp+ ' | '+tempF" small>mdi-thermometer</v-icon>
						</v-col>
						<v-col cols="1" v-if="meta.exif.FacesDetected">
							<v-icon :title="meta.exif.FacesDetected" small>mdi-face-recognition</v-icon>
						</v-col>
						<v-col cols="1"  v-if="meta.exif.SelfTimer && meta.exif.SelfTimer !== 'Off'">
							<v-icon :title="meta.exif.SelfTimer" small>mdi-camera-timer</v-icon>
						</v-col>
						<v-col cols="1" v-if="batt">
							<v-icon :title="meta.exif.BatteryLevel" small>{{ batt }}</v-icon>
						</v-col>
					</v-row>
					<div v-if="meta.exif">
						<v-row dense justify="space-between">
							<v-col cols="3">f / {{ meta.exif.Aperture }}</v-col>
							<v-col cols="auto">{{ exposure }}</v-col>
							<v-col cols="auto">ISO: {{ meta.exif.ISO }}</v-col>
						</v-row>
						<div>Focal Length: {{meta.exif.FocalLength}}</div>
						<div>Focus Mode: {{meta.exif.FocusMode}}</div>
					</div>

					<div v-if="meta.exif.LensID">{{meta.exif.LensID}}</div>
				</v-tab-item>
				<v-tab-item value="tags">
					<v-row v-if="colors" dense>
						<v-col v-for="(c,i) in colors" :key="i">
							<v-chip :color="c.toLowerCase()" small></v-chip>
						</v-col>
					</v-row>
					<tag-crumbs v-if="meta.tags && meta.tags.length" :tags="meta.tags" />
				</v-tab-item>
				<v-tab-item value="edits" v-if="meta.history && meta.history.length">
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
				</v-tab-item>
				<v-tab-item value="copy">
					<div v-if="meta.creator">Creator: {{ meta.creator }}</div>
					<div v-if="meta.rights">Rights: {{ meta.rights }}</div>
				</v-tab-item>
			</v-tabs-items>
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
import TagCrumbs from '~/components/tagCrumbs'
import Rating from '~/components/rating'
import { parseISO, format } from 'date-fns'

export default {
	props: {
		name: {},
		dir: {},
		size: {}, // in bytes
		meta: {}, //
		thumbs: {}, // full: { url: "...", width: n, height: n}
		original: {}, //{ url: "...", width: n, height: n}
	},
	data() {
		return {
			show_meta: false,
			show_history: false,
			tab: null,
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
				case 'Canon': return 'alpha-c'
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
		tempF() {
			if (!this.temp) {
				return nill
			}
			return Math.round(parseInt(this.temp.replace(/[^\d]+/g,'')) * 9/5 + 32) + "F"
		},
		focus_icon() {
			if (!this.meta || !this.meta.exif || !this.meta.exif.AFAreaMode) {
				return null
			}
			switch (this.meta.exif.AFAreaMode) {
				case 'Flexible Spot': return 'mdi-image-filter-center-focus'
				case 'Manual': return 'mdi-target'
				case 'Tracking': return 'mdi-face-recognition'
				case 'Auto': return 'mdi-focus-auto' // 'Auto' may not be actual string
			}
			return 'mdi-focus-field'
		},
		taken() {
			if (!this.meta || !this.meta.exif || !this.meta.exif.DateTimeOriginal) {
				return null
			}
			const parts = this.meta.exif.DateTimeOriginal.split(' ')
			parts[0] = parts[0].replace(/:/g, "-")

			let d = parts.join("T")
			
			if ('TimeZone' in this.meta.exif && this.meta.exif.TimeZone) {
				d += this.meta.exif.TimeZone
			} else if ('OffsetTime' in this.meta.exif && this.meta.exif.OffsetTime) {
				d += this.meta.exif.OffsetTime
			}

			return format(parseISO(d), "PPpp")
		},
		colors() {
			if (!this.meta || !this.meta.color_labels || !this.meta.color_labels.length) {
				return null
			}
			return this.meta.color_labels.map(c => {
				switch (c) {
					case "0": return "Red"
					case "1": return "Yellow"
					case "2": return "Green"
					case "3": return "Blue"
					case "4": return "Purple"
					default: return "Black"
				}
			})
		},
	},
	components: { Rating, TagCrumbs }
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