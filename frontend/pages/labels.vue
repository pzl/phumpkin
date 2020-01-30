<template>
	<v-row no-gutters style="flex-wrap: nowrap;">
		<v-navigation-drawer style="height: 90vh" mini-variant>
				<v-list-item-group v-model="selected" multiple color="primary">
					<v-list-item v-for="(c,i) in labels" :key="i" v-ripple="false" @click="unsetClear">
						<v-list-item-icon>
							<v-icon :color="toColor(c)">mdi-circle</v-icon>
						</v-list-item-icon>
					</v-list-item>
					<v-list-item v-ripple="false" @click="clear">
						<v-list-item-icon>
							<v-icon>mdi-circle-off-outline</v-icon>
						</v-list-item-icon>
					</v-list-item>
				</v-list-item-group>
		</v-navigation-drawer>
		<photo-grid style="max-width: 93%" :images="photos" @more=""></photo-grid>
	</v-row>
</template>

<script>
import PhotoGrid from '~/components/photoGrid'
import { mapState, mapMutations, mapActions } from 'vuex'


function toColor(c) {
	switch (parseInt(c)) {
		case 0: return "red";
		case 1: return "yellow";
		case 2: return "green";
		case 3: return "blue";
		case 4: return "purple";
		default: return "black"
	}
}

export default {
	data() {
		return {
			labels: [0,1,2,3,4],
			selected: [],
			photos: [],
		}
	},
	computed: {
		...mapState('images', ['err', 'loadMore']),
	},
	methods: {
		clear() {
			this.$nextTick(() => {
				this.selected = [5]
			})
		},
		unsetClear() {
			if (this.selected.indexOf(5) !== -1) {
				this.$nextTick(() => {
					this.selected.splice(this.selected.indexOf(5),1)
				})
			}
		},
		loadWithLabels(lbl) {
			if (lbl.length === 0) {
				return []
			}
			let server = location.origin
			if (server === "http://localhost:3000") {
				server = "http://localhost:6001"
			}
			this.$axios.get(server + '/api/v1/query/labels?l='+lbl.join(','))
				.then(d => { this.photos = d.data.photos })
		},
		toColor: toColor,
		...mapActions('images', ['loadImages', 'resetImages']),
	},
	watch: {
		selected(val) {
			this.loadWithLabels(val)
		}
	},
	components: { PhotoGrid }
}
</script>


<style>

</style>