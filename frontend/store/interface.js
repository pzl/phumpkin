export const state = () => ({
	view_size: false, // 1, 2, 3, 4, 6, 12
	view_dense: true,
})

export const mutations = {
	setViewSize(state, size) { state.view_size = size },
	setViewDensity(state, dense) { state.view_dense = !!dense }
}

export const actions = {
	setViewAs({ commit }, v) {
		switch (v) {
			case 'x-small':
				commit('setViewSize', 1)
				commit('setViewDensity', true)
				break
			case 'small':
				commit('setViewSize', 2)
				commit('setViewDensity', true)
				break
			case 'medium':
				commit('setViewSize', 3)
				commit('setViewDensity', true)
				break
			case 'medium-pad':
				commit('setViewSize', 3)
				commit('setViewDensity', false)
				break
			case 'large':
				commit('setViewSize', 4)
				commit('setViewDensity', true)
				break
			case 'x-large':
				commit('setViewSize', 6)
				commit('setViewDensity', false)
				break
			case 'single':
				commit('setViewSize', 12)
				commit('setViewDensity', true)
				break
			case 'auto':
			default:
				commit('setViewSize', false)
				commit('setViewDensity', true)
				break
		}
	}
}