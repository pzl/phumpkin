export const state = () => ({
	images: [],
	dirs: [],
	selected: [],
	loading: false,
	loadMore: true,
	err: false,
	sortables: [
		{ text: 'Rating', icon: 'mdi-star-half' },
		{ text: 'Date Taken', icon: 'mdi-calendar-clock' },
		{ text: 'Name', icon: 'mdi-sort-alphabetical' },
	],
	sort: 2,
	sort_asc: true,
})

export const mutations = {
	startLoading (state) { state.loading = true },
	stopLoading (state) { state.loading = false },
	errorLoading (state) { state.err = true },
	clearErr (state) { state.err = false },
	addImages (state, images) { state.images.push(...images) },
	setImages (state, images) { state.images = images },
	setDirs (state, dirs) { state.dirs = dirs },
	clearImages (state) { state.images = [] },
	clearSelection (state) { state.selected = [] },
	select (state, image) { state.selected.push(image) },
	unselect (state, image) {
		const idx = state.selected.indexOf(image)
		if (idx === -1) {
			return
		}
		state.selected.splice(idx, 1)
	},
	rate (state, { image, rating }) {
		state.images[image].xmp.rating = rating
	},
	sortBy (state, by) { state.sort = by },
	sortDir (state, dir) { state.sort_asc = dir },
	setLoadMore (state, more) { state.loadMore = more },
}

export const actions = {
	loadImages ({ commit, state }, url) {
		commit('startLoading')
		commit('clearErr')

		/*
		if (this.$sock.connected()) {
			p = this.$sock.send({
					action: "list",
					params: {
						offset: state.images.length,
						count: 10,
						sort: state.sortables[state.sort].text,
						sort_dir: state.sort_asc ? 'asc' : 'desc',
						path: state.path.join('/'),
					}
				})
		} else {
		*/
		const p = this.$fetch(url).then(d => d.data)
		//}
		p.then(d => {
			if (d.photos.length === 0) {
				commit('setLoadMore', false)
			}
			commit('addImages', d.photos)
			commit('setDirs', d.dirs)
		})
		.catch(error => {
			console.log('load image error: ')
			console.log(error)
			commit('errorLoading')
		})
		.finally(() => {
			commit('stopLoading')
		})
		return p
	},
	toggleSelect ({ commit, state }, image) {
		if (state.selected.includes(image)) {
			commit('unselect', image)
		} else {
			commit('select', image)
		}
	},
	setSelection ({ commit }, image) {
		commit('clearSelection')
		commit('select', image)
	},
	addSelection ({ commit, state }, image) {
		if (!state.selected.includes(image)) {
			commit('select', image)
		}
	},
	resetImages ({ commit }) {
		commit('setLoadMore', true)
		commit('clearSelection')
		commit('clearImages')
	},
}
