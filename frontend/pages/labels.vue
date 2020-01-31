<template>
	<v-row no-gutters style="flex-wrap: nowrap;">
		<v-navigation-drawer style="height: 90vh" mini-variant>
				<v-list-item-group v-model="selected" multiple color="primary">
					<v-list-item v-for="(c,i) in labels" :key="i" v-ripple="false">
						<v-list-item-icon>
							<v-icon :color="toColor(c)">mdi-circle</v-icon>
						</v-list-item-icon>
					</v-list-item>
				</v-list-item-group>
				<v-list-item-group :value="[]">
					<v-list-item  @click="selected=[]" v-ripple="false">
						<v-list-item-icon>
							<v-icon>mdi-circle-off-outline</v-icon>
						</v-list-item-icon>
					</v-list-item>
				</v-list-item-group>
		</v-navigation-drawer>
		<photo-grid style="max-width: 93%" :images="images" @more="loadImages(url)"></photo-grid>
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
		}
	},
	computed: {
		url(){ return "/api/v1/query/labels?l=" + this.selected.join(',') },
		...mapState('images', ['err', 'images']),
	},
	methods: {
		toColor: toColor,
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