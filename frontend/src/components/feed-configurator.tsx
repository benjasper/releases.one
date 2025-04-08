import { Link } from '@solidjs/meta'
import { Component, createEffect, createMemo, createSignal, Match, Show, Switch as SolidSwitch } from 'solid-js'
import { useConnect } from '~/context/connect-context'
import { useState } from '~/context/state-context'
import { CardDescription, CardTitle } from './ui/card'
import { Switch, SwitchControl, SwitchLabel, SwitchThumb } from './ui/switch'
import CopyText from './copy-text'
import { RepositoryStarType } from '~/lib/generated/api/v1/api_pb'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './ui/select'
import { Tooltip, TooltipContent, TooltipTrigger } from './ui/tooltip'
import { FiEye, FiInfo } from 'solid-icons/fi'
import { AiFillQuestionCircle, AiFillStar } from 'solid-icons/ai'
import StarTypeSelect from './star-type-select'

const FeedConfigurator: Component = () => {
	const connect = useConnect()
	const state = useState()
	const [prereleaseEnabled, setPrereleaseEnabled] = createSignal(true)
	const [starType, setStarType] = createSignal<RepositoryStarType | null>(null)

	state.fetchUser()

	const query = createMemo(() => {
		const params = new URLSearchParams()
		params.set('prerelease', prereleaseEnabled().toString())

		if (starType() !== null) {
			params.set('starType', starType()!.toString())
		}

		return params.toString()
	})

	const baseUrl = `${window.location.protocol}//${window.location.hostname}`;

	const rssFeed = createMemo(
		() => `${baseUrl}/rss/${state.user?.publicId ?? ''}?${query()}`
	)
	const atomFeed = createMemo(
		() => `${baseUrl}/atom/${state.user?.publicId ?? ''}?${query()}`
	)

	const setUserPublic = async (isPublic: boolean) => {
		await connect.toogleUserPublicFeed({ enabled: isPublic })
		state.fetchUser(true)
	}
	return (
		<>
			<CardTitle>Feeds</CardTitle>
			<CardDescription class="max-w-80">
				<Show when={state.user?.isPublic}>
					<Link rel="alternate" type="application/rss+xml" href={rssFeed()} />
					<Link rel="alternate" type="application/atom+xml" href={atomFeed()} />
				</Show>
				Enable your feed as a personal <b>public</b> feed. Note that everyone who has your URL has access to it
				(even if your profile is private).
			</CardDescription>
			<div class="grid gap-4">
				<div class="flex items-center space-x-4 rounded-md border p-4 justify-between">
					<div class="flex flex-col space-y-1">
						<p class="text-sm font-medium leading-none">Enabled</p>
						<p class="text-sm text-muted-foreground">Enable RSS and Atom feeds.</p>
					</div>
					<Switch checked={state.user?.isPublic ?? false} onChange={e => setUserPublic(e)}>
						<SwitchControl>
							<SwitchThumb />
						</SwitchControl>
					</Switch>
				</div>

				<Show when={state.user?.isPublic}>
					<div class="flex w-full items-center rounded-md border p-4">
						<div class="flex w-full flex-col space-y-3 justify-items-start max-w-72">
							<p class="text-sm font-medium leading-none">URLs</p>
							<p class="text-sm text-muted-foreground">Configure your feed URLs.</p>
							<StarTypeSelect starType={starType()} onChange={setStarType} />
							<Switch
								class="items-center flex gap-2 justify-between"
								checked={prereleaseEnabled()}
								onChange={setPrereleaseEnabled}>
								<SwitchLabel class="w-auto">Show prereleases</SwitchLabel>
								<SwitchControl>
									<SwitchThumb />
								</SwitchControl>
							</Switch>
							<span class="text-sm">RSS</span>
							<CopyText text={rssFeed()} />
							<span class="text-sm">Atom</span>
							<CopyText text={atomFeed()} />
						</div>
					</div>
				</Show>
			</div>
		</>
	)
}

export default FeedConfigurator
