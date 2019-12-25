<template>
	<v-card outlined  class="ma-2 mb-5 summary-card" :max-height="600">
		<v-card-title>{{name}}</v-card-title>
		<v-card-subtitle><rating :readonly="true" :value="meta.rating" /></v-card-subtitle>

		<v-card-text>
			<div>{{ sizeof }}</div>

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