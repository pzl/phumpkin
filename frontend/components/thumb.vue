<template>
	<v-hover v-slot:default="{ hover }" open-delay="50" close-delay="150">
		<v-card :class="{ 'selected': isSelected }" :raised="isSelected" >
			<v-img class="thumby" @click.stop="$emit('click', $event)" :src="image.url" v-ripple :lazy-src="lq">
				<template v-slot:placeholder>
					<v-row class="fill-height ma-0" align="center" justify="center">
						<v-progress-circular indeterminate color="grey lighten-5"></v-progress-circular>
					</v-row>
				</template>
			</v-img>

			<v-icon v-if="isSelected" class="select-check" color="success" large>mdi-checkbox-marked-circle-outline</v-icon>

			<v-menu v-model="menu" :position-x="menu_x" :position-y="menu_y" absolute offset-y >
				<v-list>
					<v-list-item v-for="n in 4" :key="n" @click="">
						<v-list-item-title>{{ n }}</v-list-item-title>
					</v-list-item>
				</v-list>
			</v-menu>

			<v-expand-transition v-if="view_size > 1 || view_size === false">
				<v-container v-if="hover" class="transition-fast-in-fast-out black darken-2 v-card--reveal white--text hidden-sm-and-down" fluid>
					<v-row dense class="d-flex justify-space-between align-center">
						<div>{{ image.name }}</div>
						<rating :value="image.rating" @input="rate({ image: index, rating: $event })" />
					</v-row>
					<div class="d-flex align-center">
						<v-tooltip bottom v-if="image.location">
							<template v-slot:activator="{ on }">
								<v-icon dark x-small v-if="image.location" v-on="on">mdi-map-marker</v-icon>
							</template>
							{{ image.location.lat }}, {{ image.location.lon }}
						</v-tooltip>
						<tags :dark="true" :tags="image.tags" />
						<v-spacer />
						<v-btn icon dark small>
							<v-icon>mdi-download</v-icon>
						</v-btn>
						<v-btn icon dark small @click="showMenu">
							<v-icon>mdi-dots-vertical</v-icon>
						</v-btn>
					</div>
				</v-container>
			</v-expand-transition>
		</v-card>
	</v-hover>
</template>

<script>
import Rating from '~/components/rating'
import Tags from '~/components/tags'
import { mapState, mapMutations, mapActions } from 'vuex'

export default {
	props: {
		index: {},
		image: {},
	},
	data() {
		return {
			lq: "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAASwAAADIBAMAAACg8cFmAAAAG1BMVEXMzMyWlpacnJyqqqrFxcWxsbGjo6O3t7e+vr6He3KoAAAACXBIWXMAAA7EAAAOxAGVKw4bAAAD0UlEQVR4nO2az2/aShDHx78wRy9pIEe7pX05Ql4r9bhu8noGVEXvaJroJUeTSjmTVqr6Z3dmfwAtrpQn2aSqvh+JXewZeb7szs4ukokAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAODP5VNxpbnrnw3XcvmluG1wSj8OPuxYvXN3REqpEfcrpY4lPl/O9r2WSg3WW6tz7pC5uloqTbEaThUHLtWVOmnQ/upM5Rurd+6QYkx9NaFooOOiIpqO6d1wz6k8Yv2jjdU7d0dffvW8opKDPjzjcaj5lv7Za56zluHG6pw7JFXcPGS0yojCY3tZ7CXXciY/YGN1zh2SyIw95DStzHhEcinfRUbOliPjVdSSUNpbnXPXSCQZjwFFx1aluctTtbRTZfKbZTmrc+6YfzhllrWJ1BttZZUDHjE7n9/ITKK3OuduKRTPVLGWaaJQikOZmfuJmoU7sXnWvNU5d4uSUmQjaRvYLbJivBpt3bhAeKtz7loWp44JonQgigIna3W0U5wiVZG3OuduZdFXjtg0Wj21LeXxklfeQUeL11vWlFuc5EcblxXvBIfNLVOxG1YiJ9cmtRKznR9uJSb/kpkXKUWpq1ur3GvZhJ4ONbfe6pw7xOwhPC+8LUoJSLZVXgqXP+Ok9ou3OucOidzm07QnTk+KZzvi6ZB7oiyoqTtBnLgzwtqYYjVZuZxfVe6OtTrnDonVNQ9C3XTe6indc1VgWttbBztv0XRwVQyo6XTKp4e+ssGLV/8x+oCn0/dc5MfUdJaXxFpaicqgD3iWp8vipeau/9H+mfnq//mY1VeaMhB7Wd7qnQEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAPB/CNznN+NJZanB619YfpKV8VWglKIBUZiRfXv8vLiRlzSvKciJihZlZfHfj5dl2uGa7p2s+KX+ws3FQge3lC7blEXv6OwFhTrlL8nzCQWLTLro+VxkLV5Qr2bLjqxRRTdOVjSzTTIJ7qm3aldWcv1+FtUlfaLLizcUXJvu8kJ+fMCmfk5vdmVlp/3cyQq1beI8CNef20xEnsRQx3l/sqjf0pjHJdCmG1MpsthEN/yxb/BaWfe92skKfJMFaXXTpixOeTMI+WmVS2iJIl3mcyujz5G8gb0drfCO9keL/spbHS15rgxJdVvR2CqRbme0wgX9ICs5pv3covmsZVmSW7So72Z0r88ltnTb3KJ09KMsc2HanZVI7RY5eTyvRHrQpaa0OJWHS7ddiZTmDbLsRJ8rX7dalvUIevVh4z2Su6cW0Ehw+tQKAAB/DN8B6Jikju6t6uAAAAAASUVORK5CYII=",
			hover_reject: false,
			menu: false,
			menu_x: 0,
			menu_y: 0,
		}
	},
	computed: {
		isSelected(){
			return this.selected.includes(this.index)
		},
		rating() {
			return this.hover_reject ? 0 : this.image.rating
		},
		...mapState('images', ['selected']),
		...mapState('interface', ['view_size']),
	},
	methods: {
		showMenu(e) {
			e.preventDefault()

			this.menu = false
			this.menu_x = e.clientX
			this.menu_y = e.clientY
			this.$nextTick(() => {
				this.menu = true
			})

		},
		...mapMutations('images', ['rate']),
	},
	components: { Rating, Tags },
}
</script>

<style>
.select-check {
	position: absolute;
	top: 2px;
	right: 2px;
	background: rgba(255,255,255,0.2);
	box-shadow: 0 0 2px 0 rgba(255,255,255,0.6);
	border-radius: 50% !important;
}
</style>