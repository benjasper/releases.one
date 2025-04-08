import { A } from '@solidjs/router'
import { ImGithub } from 'solid-icons/im'
import { Component } from 'solid-js'
import { buttonVariants } from '~/components/ui/button'

const LandingPage: Component = () => {
	return (
		<div class="flex flex-col gap-16 py-8 justify-between h-svh">
			<section class="flex container justify-between items-center">
				<h1 class="text-3xl">releases.one</h1>
				<A class={buttonVariants({ variant: 'default' })} href="/login">
					Login
				</A>
			</section>

			<section class="flex flex-col gap-4 container justify-center items-center">
				<h2 class="text-4xl">Your starred GitHub repositories as a feed</h2>
				<p class="text-muted-foreground text-lg mb-4 text-center">
					releases.one is a free and open-source tool
					<br /> to help you keep track of your starred and watched GitHub repositories by providing a feed of
					your latest releases.
				</p>
				<A class={buttonVariants({ variant: 'default' })} href="/login">
					Get my feed URL
				</A>
			</section>

			{/* Footer */}
			<footer class="flex flex-col gap-4 container justify-center items-center">
				<p class="text-muted-foreground text-sm">
					releases.one is an open-source project by{' '}
					<a href="https://benjaminjasper.com" class="hover:underline" target="_blank" rel="noopener noreferrer">
						Benjamin Jasper
					</a>
					.
				</p>
				<a href="https://github.com/benjasper/releases.one" class="flex items-center gap-2" target="_blank">
					<ImGithub />
				</a>
			</footer>
		</div>
	)
}

export default LandingPage
