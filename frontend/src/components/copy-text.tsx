import { Component, createSignal, Show } from 'solid-js'
import { Button } from './ui/button'
import { FiCheck, FiCopy } from 'solid-icons/fi'

type CopyTextProps = {
	text: string
}

const CopyText: Component<CopyTextProps> = props => {
	const [copied, setCopied] = createSignal(false)

	const copy = () => {
		navigator.clipboard.writeText(props.text)
		setCopied(true)
		setTimeout(() => setCopied(false), 3000)
	}

	return (
		<div class="flex gap-2">
			<span class="flex-1 select-all whitespace-nowrap overflow-auto no-scrollbar w-full text-sm p-2 text-muted-foreground border rounded-md">
				{props.text}
			</span>
			<Button class="cursor-pointer" onClick={copy}>
				<Show when={copied()} fallback={<FiCopy class="w-3 h-3" />}>
					<FiCheck class="w-3 h-3" />
				</Show>
			</Button>
		</div>
	)
}

export default CopyText
