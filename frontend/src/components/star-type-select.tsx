import { Component, createSignal, Match, Switch } from 'solid-js'
import { RepositoryStarType } from '~/lib/generated/api/v1/api_pb'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './ui/select'
import { FiEye } from 'solid-icons/fi'
import { AiFillQuestionCircle, AiFillStar } from 'solid-icons/ai'
import { Tooltip, TooltipContent, TooltipTrigger } from './ui/tooltip'

type NullableStarType = null | RepositoryStarType

type Props = {
	starType: NullableStarType
	onChange: (starType: NullableStarType) => void
}

const StarTypeSelect: Component<Props> = props => {
	return (
		<div class="flex items-center gap-4 justify-between">
			<label for="star-type" class="flex items-center text-sm font-medium leading-none">
				Repository type
				<Tooltip>
					<TooltipTrigger>
						<AiFillQuestionCircle class="ml-1 h-4 w-4 cursor-help text-muted-foreground" />
					</TooltipTrigger>
					<TooltipContent>
						<p class="text-sm text-muted-foreground max-w-72">
							You can choose to receive notifications for both watched and starred repositories, just
							watched repositories, or just starred repositories.
						</p>
					</TooltipContent>
				</Tooltip>
			</label>
			<Select
				id="star-type"
				value={props.starType}
				onChange={value => props.onChange(value)}
				options={[null, RepositoryStarType.STAR, RepositoryStarType.WATCH]}
				placeholder={<StarTypeLabel starType={null}></StarTypeLabel>}
				itemComponent={props => (
					<SelectItem item={props.item} class="cursor-pointer">
						<StarTypeLabel starType={props.item.rawValue} />
					</SelectItem>
				)}>
				<SelectTrigger aria-label="Type" class="cursor-pointer">
					<SelectValue<RepositoryStarType | null>>
						{state => <StarTypeLabel starType={state.selectedOption()}></StarTypeLabel>}
					</SelectValue>
				</SelectTrigger>
				<SelectContent />
			</Select>
		</div>
	)
}

const StarTypeLabel: Component<{ starType: RepositoryStarType | null }> = props => {
	return (
		<div class="flex gap-1 items-center">
			<Switch>
				<Match when={props.starType === null}>Both</Match>
				<Match when={props.starType === RepositoryStarType.STAR}>
					<AiFillStar class="w-4" /> Starred
				</Match>
				<Match when={props.starType === RepositoryStarType.WATCH}>
					<FiEye class="w-4" /> Watched
				</Match>
			</Switch>
		</div>
	)
}

export default StarTypeSelect
