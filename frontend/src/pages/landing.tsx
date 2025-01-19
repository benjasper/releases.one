import { A } from '@solidjs/router'
import { Component } from 'solid-js'
import { buttonVariants } from '~/components/ui/button'

const LandingPage: Component = () => {
	return (
		<div class="flex flex-col gap-16 py-8 justify-center">
			<section class="flex container justify-between items-center">
				<h1 class="text-3xl">releases.one</h1>
				<A class={buttonVariants({ variant: 'default' })} href="/login">
					Login
				</A>
			</section>

			<section class="flex flex-col gap-4 container justify-center items-center">
				<h2 class="text-4xl">Your starred GitHub repositories as a feed</h2>
				<p class="text-muted-foreground text-lg mb-4">
					releases.one is a free and open-source tool to help you keep track of your starred GitHub repositories by providing a
					feed of your latest releases.
				</p>
				<A class={buttonVariants({ variant: 'default' })} href="/login">
					Get my feed URL
				</A>
			</section>
		</div>
	)
}

export default LandingPage
