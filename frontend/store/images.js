import Vue from 'vue'

export const state = () => ({
	images:[],
	selected: [],
	loading: false,
	sort: 'name',
	sort_asc: true,
})

export const mutations = {
	startLoading (state) { state.loading = true },
	stopLoading (state) { state.loading = false },
	addImages(state, images) { state.images.push(...images) },
	setImages(state,images) { state.images = images },
	clearImages(state) { state.images = [] },
	clearSelection (state) { state.selected = [] },
	select (state, image) { state.selected.push(image) },
	unselect (state, image) {
		const idx = state.selected.indexOf(image)
		if (idx === -1) {
			return
		}
		state.selected.splice(idx, 1)
	},
	rate(state, { image, rating }) {
		state.images[image].xmp.rating = rating
	},
	sortBy(state, by) { state.sort = by },
	sortDir(state, dir){ state.sort_asc = dir },
}

export const actions = {
	async loadImages({ commit, state }) {
		commit('startLoading')

		if (this.$sock.connected()) {
			const cmd = {
				action:"list",
				params: {
					offset: state.images.length,
					count: 10,
					sort: state.sort,
					sort_dir: state.sort_asc ? 'asc' : 'desc',
				}
			}
			console.log('sending load cmd: ',cmd)

			this.$sock.send(cmd)
				.then(data => {
					commit('addImages', data)
				})
				.catch(error => {
					console.log('sock error: ')
					console.log(error)
				})
				.finally(() => {
					commit('stopLoading')
				})
		} else {
			// fall back to HTTP
			try {
				// @todo: replace with location.origin + "/api.."
				const response = await this.$axios.$get("http://localhost:6001/api/v1/photos?count=30&offset="+(state.images.length||0)+"&sort="+state.sort+"&sort_dir="+(state.sort_asc ? 'asc' : 'desc'))
				commit('addImages', response.photos)
			} catch (error) {
				// oh no
				console.log("http error: ")
				console.log(error)
			}
			commit('stopLoading')
		}
	},
	toggleSelect({ commit, state }, image) {
		if (state.selected.includes(image)) {
			commit('unselect', image)
		} else {
			commit('select', image)
		}
	},
	setSelection({ commit }, image) {
		commit('clearSelection')
		commit('select', image)
	},
	addSelection({ commit, state }, image) {
		if (!state.selected.includes(image)) {
			commit('select', image)
		}
	},
	resetImages({ commit }) {
		commit('clearSelection')
		commit('clearImages')
	},
}