import {
	Component,
	createEffect,
	createMemo,
	createResource,
	createSignal,
	For,
	onCleanup,
	onMount,
	Show,
} from 'solid-js'
import { useConnect } from '~/context/connect-context'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '~/components/ui/card'
import { Avatar, AvatarFallback } from '~/components/ui/avatar'
import { timestampDate } from '@bufbuild/protobuf/wkt'
import { Button, buttonVariants } from '~/components/ui/button'
import { Popover, PopoverContent, PopoverTrigger } from '~/components/ui/popover'
import { Switch, SwitchControl, SwitchDescription, SwitchLabel, SwitchThumb } from '~/components/ui/switch'
import { TextField, TextFieldInput } from '~/components/ui/text-field'
import { Badge } from '~/components/ui/badge'
import CopyText from '~/components/copy-text'
import DarkModeToggle from '~/components/dark-mode-toggle'
import { FiArrowUp, FiExternalLink, FiRss } from 'solid-icons/fi'
import { formatDistance } from 'date-fns/formatDistance'
import { Tooltip, TooltipContent, TooltipTrigger } from '~/components/ui/tooltip'
import { Link } from '@solidjs/meta'

const TimelinePage: Component = () => {
	const connect = useConnect()
	const [timeline] = createResource(() => connect.getRepositories({}))
	const [user, { refetch: refetchUser }] = createResource(() => connect.getMyUser({}))

	const [search, setSearch] = createSignal('')
	const [descriptionEnabled, setDescriptionEnabled] = createSignal(true)
	const [isScrollingDown, setIsScrollingDown] = createSignal(false)
	const [now, setNow] = createSignal(Date.now())

	const filteredTimeline = () =>
		timeline()?.timeline.filter(x => x.repositoryName.toLowerCase().includes(search().toLowerCase())) ?? []

	const rssFeed = createMemo(() => `${import.meta.env.VITE_API_BASE_URL}/rss/${user()?.publicId ?? ''}`)
	const atomFeed = createMemo(() => `${import.meta.env.VITE_API_BASE_URL}/atom/${user()?.publicId ?? ''}`)

	const setUserPublic = async (isPublic: boolean) => {
		await connect.toogleUserPublicFeed({ enabled: isPublic })
		refetchUser()
	}

	// Signal when scrolling down
	const handleScroll = () => {
		setIsScrollingDown(window.scrollY > 20)
	}

	onMount(() => {
		window.addEventListener('scroll', handleScroll)
	})

	onCleanup(() => {
		window.removeEventListener('scroll', handleScroll)
	})

	setInterval(() => {
		setNow(Date.now())
	}, 1000)

	const calculateDuration = (date: Date): string => {
		return formatDistance(date, now(), { addSuffix: true })
	}

	return (
		<>
			<Show when={user()?.isPublic}>
				<Link rel="alternate" type="application/rss+xml" href={rssFeed()} />
				<Link rel="alternate" type="application/atom+xml" href={atomFeed()} />
			</Show>

			<div class="flex flex-col gap-4 container pt-4">
				<Button
					classList={{
						'opacity-100': isScrollingDown(),
					}}
					class="fixed bottom-5 right-5 z-50 p-3 opacity-0 transition-all cursor-pointer"
					onClick={() => window.scrollTo({ top: 0, behavior: 'smooth' })}>
					<FiArrowUp class="w-6 h-6" />
				</Button>
				<div class="flex items-center justify-between space-y-2">
					<div>
						<h2 class="text-2xl font-bold tracking-tight">Welcome back!</h2>
						<p class="text-muted-foreground">Here&apos;s a list of recent releases!</p>
					</div>
					<div class="flex items-center space-x-4">
						<DarkModeToggle />
						<Popover>
							<PopoverTrigger as={Button<'button'>} class="cursor-pointer">
								<FiRss class="w-6 h-6" />
								Feeds
							</PopoverTrigger>
							<PopoverContent class="flex flex-col gap-4 !w-auto">
								<CardTitle>Feeds</CardTitle>
								<CardDescription class="max-w-80">
									Enable your feed as a personal <b>public</b> feed. Note that everyone who has the
									URL can has access to it (even if your profile is private).
								</CardDescription>
								<div class="grid gap-4">
									<div class="flex items-center space-x-4 rounded-md border p-4">
										<div class="flex flex-col space-y-1">
											<p class="text-sm font-medium leading-none">Enabled</p>
											<p class="text-sm text-muted-foreground">
												Get this feed as a RSS or Atom feed.
											</p>
										</div>
										<Switch checked={user()?.isPublic ?? false} onChange={e => setUserPublic(e)}>
											<SwitchControl>
												<SwitchThumb />
											</SwitchControl>
										</Switch>
									</div>

									<Show when={user()?.isPublic}>
										<div class="flex w-full items-center rounded-md border p-4">
											<div class="flex w-full flex-col space-y-2 justify-items-start max-w-72">
												<p class="text-sm font-medium leading-none">URLs</p>
												<p class="text-sm text-muted-foreground">
													Get this feed as a RSS or Atom feed.
												</p>
												<span class="text-sm">RSS</span>
												<CopyText text={rssFeed()} />
												<span class="text-sm">Atom</span>
												<CopyText text={atomFeed()} />
											</div>
										</div>
									</Show>
								</div>
							</PopoverContent>
						</Popover>
						<Avatar class="size-10">
							<AvatarFallback class="uppercase">{user()?.name[0]}</AvatarFallback>
						</Avatar>
					</div>
				</div>
				<div class="flex gap-4 items-center justify-between md:justify-start">
					<TextField>
						<TextFieldInput
							placeholder={'Search repositories'}
							value={search()}
							onInput={e => setSearch(e.currentTarget.value)}
						/>
					</TextField>

					<Switch
						class="items-center flex gap-2"
						checked={descriptionEnabled()}
						onChange={setDescriptionEnabled}>
						<SwitchControl>
							<SwitchThumb />
						</SwitchControl>
						<SwitchLabel class="w-auto">Show release description</SwitchLabel>
					</Switch>
				</div>
				<For each={filteredTimeline()}>
					{timelineItem => (
						<Card class="mx-auto w-full max-w-120 hover:shadow-lg transition-shadow duration-200">
							<CardHeader class="!p-0">
								<img
									class="rounded-t-lg aspect-2/1"
									src={timelineItem.imageUrl}
									loading="lazy"
									alt={timelineItem.name}
								/>
							</CardHeader>
							<CardContent class="flex flex-col !pb-0 pt-4 prose dark:prose-invert">
								<a
									href={timelineItem.repositoryUrl}
									class="no-underline hover:underline"
									target="_blank"
									rel="noopener noreferrer">
									<span class="font-normal">{timelineItem.repositoryName}</span>
								</a>
								<a
									href={timelineItem.url}
									class="no-underline hover:underline"
									target="_blank"
									rel="noopener noreferrer">
									<h2 class="!mt-0 !mb-4 font-normal">{timelineItem.name}</h2>
								</a>
								<Show when={descriptionEnabled()}>
									<div class="prose-sm" innerHTML={timelineItem.description}></div>
								</Show>
							</CardContent>
							<CardFooter class="flex justify-between text-muted-foreground !pt-2 text-sm">
								<Tooltip>
									<TooltipTrigger>
										{calculateDuration(timestampDate(timelineItem.releasedAt!))}
									</TooltipTrigger>
									<TooltipContent>
										{timestampDate(timelineItem.releasedAt!).toLocaleString()}
									</TooltipContent>
								</Tooltip>

								<a
									href={timelineItem.url}
									target="_blank"
									rel="noopener noreferrer"
									class={buttonVariants({ variant: 'ghost', size: 'icon' })}>
									<FiExternalLink class="w-4 h-4" />
								</a>
							</CardFooter>
						</Card>
					)}
				</For>
			</div>
		</>
	)
}

export default TimelinePage
