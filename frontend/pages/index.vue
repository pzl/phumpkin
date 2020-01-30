<template>
	<photo-grid :images="images" @more="loadImages">
		<template v-slot:before>
			<v-row class="content-group">
				<path-crumbs />
			</v-row>
			<v-row class="content-group">
				<directory v-for="(d,i) in dirs" :key="d+i" :name="d" @click="onDirClick(d, $event)" />
			</v-row>
		</template>
	</photo-grid>
</template>

<script>
import Directory from '~/components/directory'
import PathCrumbs from '~/components/pathCrumbs'
import PhotoGrid from '~/components/photoGrid'
import { mapState, mapMutations, mapActions } from 'vuex'

export default {
	data() {
		return {
		}
	},
	computed: {
		...mapState('images', ['images', 'dirs', 'selected', 'loading', 'err', 'loadMore']),
	},
	methods: {
		onDirClick(dir, e) {
			e.preventDefault()

			this.pushPath(dir)
			this.resetImages()
			this.loadImages()
		},
		...mapMutations('images', ['pushPath']),
		...mapActions('images', ['loadImages', 'resetImages']),
	},
	mounted() {
		this.loadImages()
	},
	components: { PhotoGrid, Directory, PathCrumbs }
}
</script>


<style>
.thumby {
	cursor: pointer;
}

.v-card--reveal {
	align-items: center;
  opacity: .5;
  bottom: 0;
  position: absolute;
  width: 100%;

  user-select: none; /* prevent name highlighting on shift-click */
}


.content-group {
	max-width: 100%;
}
</style>