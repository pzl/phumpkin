<template>
	<v-row no-gutters style="flex-wrap: nowrap;">
		<photo-grid :images="images" @more="loadImages(url)"></photo-grid>
	</v-row>
</template>

<script>
import PhotoGrid from '~/components/photoGrid'
import { mapState, mapMutations, mapActions } from 'vuex'

export default {
	data() {
		return {
			
		}
	},
	computed: {
		url(){ return "/api/v1/query/faces" },
		...mapState('images', ['err', 'images']),
	},
	methods: {
		...mapActions('images', ['loadImages', 'resetImages']),
	},
	mounted() {
		this.resetImages()
	},
	components: { PhotoGrid }
}
</script>


<style>

</style>