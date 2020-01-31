<template>
	<photo-grid :images="images" @more="loadImages(url)">
		<template v-slot:before>
			<v-row class="content-group">
				<path-crumbs :path="path" @clear="clearPath" @pop="popPath" />
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
			path: []
		}
	},
	computed: {
		url() {
			return "/api/v1/photos/?path=" + this.path.join('/')
		},
		...mapState('images', ['images', 'dirs', 'selected', 'loading', 'err', 'loadMore']),
	},
	methods: {
		pushPath(d) { this.path.push(d) },
		clearPath () { this.path = [] },
		popPath () { return this.path.pop() },
		onDirClick(dir, e) {
			e.preventDefault()
			this.pushPath(dir)
		},
		...mapActions('images', ['loadImages', 'resetImages']),
	},
	watch: {
		path() {
			this.resetImages()
		}
	},
	mounted() {
		this.loadImages(this.url)
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