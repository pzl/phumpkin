<template>
	<v-row no-gutters style="flex-wrap: nowrap;">
		<v-navigation-drawer style="height: 90vh; flex-shrink: 0" mini-variant>
				<v-list-item-group v-model="selected" multiple color="primary">
					<v-list-item v-for="r in ratings" :key="r" v-ripple="false" :value="r">
						<v-list-item-icon>
							<template v-if="r > 0">
								<v-icon v-for="i in r" :key="r + '-'+ i" :size="size(r)">{{ icon(r) }}</v-icon>
							</template>
							<v-icon v-else :color="r < 0 ? 'red darken-3' : undefined">{{ icon(r) }}</v-icon>
						</v-list-item-icon>
					</v-list-item>
				</v-list-item-group>
		</v-navigation-drawer>
		<photo-grid :images="images" @more="loadImages(url)"></photo-grid>
	</v-row>
</template>

<script>
import PhotoGrid from '~/components/photoGrid'
import { mapState, mapMutations, mapActions } from 'vuex'

export default {
	data() {
		return {
			ratings: [-1,0,1,2,3,4,5].reverse(),
			selected: [],
		}
	},
	computed: {
		url(){ return "/api/v1/query/rating?r=" + this.selected.join(',') },
		...mapState('images', ['err', 'images']),
	},
	methods: {
		icon(v) {
			switch (v) {
				case -1: return 'mdi-cancel'
				case 0: return 'mdi-star-outline'
				default: return 'mdi-star'
			}
		},
		size(v) {
			switch (v) {
				case 5: return 10
				case 4: return 13
				case 3: return 17
				default: return 20
			}
		},
		...mapActions('images', ['loadImages', 'resetImages']),
	},
	watch: {
		selected() {
			this.resetImages()
		}
	},
	mounted() {
		this.resetImages()
	},
	components: { PhotoGrid }
}
</script>


<style>

</style>