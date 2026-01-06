From: https://platform.openai.com/docs/guides/retrieval?attributes-filter-example=filename&vector-store-operations=list 
- diferenca entre  busca semantica e textual
- poder de rescrever uma query

From:RAG From Scratch: Part 1  https://www.youtube.com/watch?v=wd7TZ4w1mSw&list=PLfaIDFEXuae2LXbO1_PKyVJiQ23ZztA0x&index=1
- llm e como se fosse o CPU de um kernel semantico
- Retrival pipeline
  - QUESTION  -> Index(document) -> Prompt(question with Document) -> llm -> ANSWER

from: RAG From Scratch: Part 2 https://www.youtube.com/watch?v=bjb_EMsTDKI&list=PLfaIDFEXuae2LXbO1_PKyVJiQ23ZztA0x&index=2
- there ar different ways to create the embed vector, it is an matematical issue
- will be important register token length of all steps

from: meet com IgorTiburcio
- um rag para cada intencao
- pipeline pode gerar n rags para 1 source a depender da necessidade
- resumir o source com uma ia barata, embedar o resumo,funciona melhor em algumas situacoes do que embedar o texto original
- podemos criar um rag para personagens perspective(fala, pensamento, medos, situacoes)
- podemos embedar documentos q ensinem pra ia o q é plot, como sao mundos, etc... essa é a parte q direito autoral nao cobre, e eh complicado expor, talvez fazer um modo do usuario colocar sua definicao de plot e arco para nao cair em problemas
