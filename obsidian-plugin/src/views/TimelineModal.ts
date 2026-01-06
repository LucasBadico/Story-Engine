import { Modal, App, Notice } from "obsidian";
// @ts-ignore - d3 types may not resolve correctly in all environments
import * as d3 from "d3";
import { WorldEvent, TimeConfig } from "../types";

interface TimelineEvent extends WorldEvent {
	timeline_position: number;
	parent_id: string | null;
	is_epoch: boolean;
}

export class TimelineModal extends Modal {
	events: WorldEvent[];
	timeConfig: TimeConfig | null;
	
	constructor(app: App, events: WorldEvent[], timeConfig: TimeConfig | null) {
		super(app);
		this.events = events;
		this.timeConfig = timeConfig;
	}
	
	onOpen() {
		const { contentEl } = this;
		contentEl.empty();
		contentEl.addClass("story-engine-timeline-modal");
		
		contentEl.createEl("h2", { text: "World Timeline" });
		
		// Container para o grafico D3
		const chartContainer = contentEl.createDiv({ cls: "story-engine-timeline-chart" });
		
		this.renderTimeline(chartContainer);
	}
	
	renderTimeline(container: HTMLElement) {
		if (this.events.length === 0) {
			container.createEl("p", { text: "No events with timeline data found." });
			return;
		}

		// Dimensoes
		const width = 800;
		const height = 400;
		const margin = { top: 20, right: 30, bottom: 50, left: 60 };
		
		// Criar SVG
		const svg = d3.select(container)
			.append("svg")
			.attr("width", width)
			.attr("height", height);
		
		// Filtrar eventos com timeline_position
		const eventsWithPosition: TimelineEvent[] = this.events.map((e, i) => ({
			...e,
			timeline_position: e.timeline_position ?? i * 10,
			parent_id: e.parent_id ?? null,
			is_epoch: e.is_epoch ?? false,
		}));

		// Escala X (timeline_position em anos)
		const xExtent = d3.extent(eventsWithPosition, (d: TimelineEvent) => d.timeline_position) as [number, number];
		const xScale = d3.scaleLinear()
			.domain([xExtent[0] - 1, xExtent[1] + 1])
			.range([margin.left, width - margin.right]);
		
		// Escala Y (importance 1-10)
		const yScale = d3.scaleLinear()
			.domain([0, 10])
			.range([height - margin.bottom, margin.top]);
		
		// Escala de cor por tipo de evento
		const colorScale = d3.scaleOrdinal(d3.schemeCategory10);
		
		// Eixo X
		svg.append("g")
			.attr("transform", `translate(0,${height - margin.bottom})`)
			.call(d3.axisBottom(xScale).tickFormat((d: d3.NumberValue) => `Year ${d}`));
		
		// Eixo Y
		svg.append("g")
			.attr("transform", `translate(${margin.left},0)`)
			.call(d3.axisLeft(yScale).tickFormat((d: d3.NumberValue) => `Imp ${d}`));
		
		// Conectar eventos pai-filho com linhas (se houver parent_id)
		const eventsWithParent = eventsWithPosition.filter(e => e.parent_id);
		svg.selectAll("line.parent-link")
			.data(eventsWithParent)
			.join("line")
			.attr("class", "parent-link")
			.attr("x1", (d: TimelineEvent) => {
				const parent = eventsWithPosition.find(e => e.id === d.parent_id);
				return parent ? xScale(parent.timeline_position) : 0;
			})
			.attr("y1", (d: TimelineEvent) => {
				const parent = eventsWithPosition.find(e => e.id === d.parent_id);
				return parent ? yScale(parent.importance) : 0;
			})
			.attr("x2", (d: TimelineEvent) => xScale(d.timeline_position))
			.attr("y2", (d: TimelineEvent) => yScale(d.importance))
			.attr("stroke", "#999")
			.attr("stroke-dasharray", "4,2")
			.attr("stroke-width", 1);
		
		// Plotar eventos como circulos
		svg.selectAll("circle")
			.data(eventsWithPosition)
			.join("circle")
			.attr("cx", (d: TimelineEvent) => xScale(d.timeline_position))
			.attr("cy", (d: TimelineEvent) => yScale(d.importance))
			.attr("r", (d: TimelineEvent) => d.is_epoch ? 10 : 6)
			.attr("fill", (d: TimelineEvent) => colorScale(d.type || "default"))
			.attr("stroke", (d: TimelineEvent) => d.is_epoch ? "gold" : "none")
			.attr("stroke-width", 3)
			.style("cursor", "pointer")
			.on("click", (_event: MouseEvent, d: TimelineEvent) => this.showEventDetails(d))
			.append("title")
			.text((d: TimelineEvent) => `${d.name} (Year ${d.timeline_position})`);
		
		// Labels para eventos importantes (importance >= 8)
		svg.selectAll("text.event-label")
			.data(eventsWithPosition.filter(e => e.importance >= 8))
			.join("text")
			.attr("class", "event-label")
			.attr("x", (d: TimelineEvent) => xScale(d.timeline_position))
			.attr("y", (d: TimelineEvent) => yScale(d.importance) - 12)
			.attr("text-anchor", "middle")
			.attr("font-size", "10px")
			.text((d: TimelineEvent) => d.name.substring(0, 15) + (d.name.length > 15 ? "..." : ""));
	}
	
	showEventDetails(event: WorldEvent) {
		// Mostrar detalhes do evento em um popup ou redirecionar
		new Notice(`Event: ${event.name}\nYear: ${event.timeline_position ?? "N/A"}\nImportance: ${event.importance}`);
	}
	
	onClose() {
		const { contentEl } = this;
		contentEl.empty();
	}
}

