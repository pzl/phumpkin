export const state = () => ({
	images: [...Array(50).keys()].map(i => {
		return {
			name: `IMG_00${i+1 < 10 ? '0'+(i+1) : i+1}.JPG`,
			url: `http://picsum.photos/300/200?random=${i}`,
			rating: Math.floor(Math.random() * 7)-1, // -1 -> 5
			tags: ["Nature", "Door", "Switzerland", "People", "Animal"].filter(t => Math.random() > 0.65),
			location: Math.random() > 0.8 ? { lat: "42,58.824683N", lon: "78,51.318550W" } : null,
		}
	}),
	selected: []
})

export const mutations = {
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
		state.images[image].rating = rating
	}
}

export const actions = {
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
	}
}