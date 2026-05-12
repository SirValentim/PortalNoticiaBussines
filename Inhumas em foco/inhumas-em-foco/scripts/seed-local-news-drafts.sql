INSERT OR IGNORE INTO categories (slug, name, description, requires_editorial_notes)
VALUES ('historia-de-inhumas', 'Historia de Inhumas', 'Memoria, curiosidades e fatos historicos de Inhumas GO.', false);

INSERT OR IGNORE INTO posts (
  title, slug, excerpt, content, cover_image_key, category_id, author_id, status,
  is_sponsored, editorial_notes, editor_responsible, published_at, publish_at
)
VALUES
(
  'Quem foi o primeiro prefeito de Inhumas?',
  'quem-foi-o-primeiro-prefeito-de-inhumas',
  'Jose Rodrigues Rabelo aparece em fontes historicas como o primeiro prefeito de Inhumas, nomeado no contexto da emancipacao municipal de 1931.',
  '<p>Uma das perguntas mais buscadas por quem pesquisa a historia politica de Inhumas e simples: quem foi o primeiro prefeito da cidade?</p><p>De acordo com pesquisa academica da PUC Goias que trata da memoria politica local, o Decreto nº 602, de 19 de janeiro de 1931, emancipou Inhumas e nomeou Jose Rodrigues Rabelo como prefeito provisorio. A mesma pesquisa registra que ele exerceu o cargo a partir da instalacao do municipio, em 19 de marco de 1931, seguindo ate o fim de 1935.</p><h2>O contexto da emancipacao</h2><p>O processo de emancipacao de Inhumas ocorreu em meio a articulacoes politicas locais e estaduais no inicio da decada de 1930. A pesquisa cita a participacao de liderancas locais e o pedido levado ao governo de Goias, entao sob Pedro Ludovico.</p><h2>Primeiras medidas da administracao</h2><p>Entre os registros associados ao primeiro governo municipal estao a organizacao de logradouros publicos e iniciativas ligadas a educacao, como a criacao do Grupo Escolar 19 de Marco.</p><h2>Por que isso importa?</h2><p>Conhecer o primeiro prefeito ajuda a entender como Inhumas deixou de ser distrito e passou a construir sua propria administracao municipal. Tambem ajuda moradores, estudantes e pesquisadores a localizar nomes e datas importantes da cidade.</p><h2>Fontes consultadas</h2><ul><li><a href="https://tede2.pucgoias.edu.br/bitstream/tede/2266/1/EBER%20DORNELAS%20DA%20COSTA%20SOUZA.pdf">Dissertacao PUC Goias - Eber Dornelas da Costa Souza</a></li><li><a href="https://inhumas.go.leg.br/historia-da-camara/">Camara Municipal de Inhumas - Historia da Camara</a></li></ul>',
  '',
  (SELECT id FROM categories WHERE slug = 'historia-de-inhumas'),
  NULL,
  'draft',
  false,
  'Texto historico baseado em fonte academica e fonte institucional. Revisar antes de publicar.',
  'Redacao Inhumas em Foco',
  NULL,
  NULL
),
(
  'Inhumas em numeros: populacao, area e dados oficiais do IBGE',
  'inhumas-em-numeros-populacao-area-e-dados-oficiais',
  'Dados do IBGE mostram Inhumas com mais de 52 mil moradores no Censo 2022 e populacao estimada acima de 53 mil pessoas em 2025.',
  '<p>Inhumas e uma das cidades mais conhecidas da regiao metropolitana de Goiania e aparece nos dados oficiais do IBGE com indicadores importantes para moradores, empresas e gestores publicos.</p><h2>Populacao e territorio</h2><p>Segundo a pagina Cidades e Estados do IBGE, Inhumas tinha 52.204 pessoas no Censo 2022. A estimativa populacional informada para 2025 e de 53.884 pessoas. A area territorial registrada para 2025 e de 614,889 km².</p><h2>Densidade e indicadores</h2><p>O mesmo painel do IBGE informa densidade demografica de 84,90 habitantes por quilometro quadrado em 2022. O gentílico oficial e inhumense.</p><h2>Economia e qualidade de vida</h2><p>O IBGE tambem apresenta PIB per capita de R$ 36.024,89 para 2023 e IDHM de 0,720, indicador referente a 2010. Esses numeros ajudam a contextualizar o tamanho e o perfil socioeconomico do municipio.</p><h2>Por que acompanhar esses dados?</h2><p>Dados oficiais sao importantes para planejamento urbano, comercio local, politicas publicas e tambem para quem deseja investir ou empreender em Inhumas.</p><h2>Fonte consultada</h2><ul><li><a href="https://www.ibge.gov.br/cidades-e-estados/go/inhumas.html">IBGE Cidades e Estados - Inhumas GO</a></li></ul>',
  '',
  (SELECT id FROM categories WHERE slug = 'noticias'),
  NULL,
  'draft',
  false,
  'Texto de servico com dados oficiais do IBGE. Conferir numeros antes da publicacao final.',
  'Redacao Inhumas em Foco',
  NULL,
  NULL
),
(
  'A Camara Municipal de Inhumas e a redemocratizacao de 1947',
  'camara-municipal-de-inhumas-redemocratizacao-1947',
  'A historia institucional da Camara de Inhumas registra a instalacao do Legislativo local em 1947, apos o fim do Estado Novo.',
  '<p>A Camara Municipal de Inhumas tem papel central na historia politica da cidade. Segundo a propria Casa Legislativa, o Poder Legislativo inhumense foi instalado em 1947, em um periodo de redemocratizacao do Brasil apos o fim do Estado Novo.</p><h2>O que faz a Camara?</h2><p>A Camara e o espaco onde vereadores apresentam projetos de lei, requerimentos, fiscalizam a administracao municipal e participam de decisoes que afetam a vida da cidade.</p><h2>Comissoes e Mesa Diretora</h2><p>A pagina institucional da Camara explica que os trabalhos legislativos sao organizados por Mesa Diretora e comissoes. Esses grupos analisam temas como legislacao, financas, administracao municipal, educacao, saude, obras e patrimonio.</p><h2>Por que esse marco e importante?</h2><p>Entender a instalacao do Legislativo ajuda a perceber como Inhumas passou a organizar sua representacao politica local no periodo democratico posterior a ditadura varguista.</p><h2>Fonte consultada</h2><ul><li><a href="https://inhumas.go.leg.br/historia-da-camara/">Camara Municipal de Inhumas - Historia da Camara</a></li></ul>',
  '',
  (SELECT id FROM categories WHERE slug = 'historia-de-inhumas'),
  NULL,
  'draft',
  false,
  'Texto institucional explicativo baseado em fonte da Camara Municipal.',
  'Redacao Inhumas em Foco',
  NULL,
  NULL
),
(
  'Por que 19 de marco e uma data importante para Inhumas?',
  'por-que-19-de-marco-e-importante-para-inhumas',
  'A data remete a instalacao do municipio de Inhumas em 1931, marco decisivo para a administracao propria da cidade.',
  '<p>Para quem pesquisa a historia de Inhumas, 19 de marco aparece como uma data de grande valor simbolico e administrativo.</p><h2>A instalacao do municipio</h2><p>Pesquisa academica da PUC Goias registra que, apos o Decreto nº 602 de 19 de janeiro de 1931, Inhumas foi instalada como municipio em 19 de marco daquele ano. A partir dali, a cidade passou a organizar sua administracao propria.</p><h2>Emancipacao e memoria local</h2><p>A emancipacao nao foi apenas uma mudanca burocratica. Ela consolidou a autonomia politica do territorio e marcou o inicio de uma nova etapa na historia local, com prefeito, paço municipal e estrutura administrativa.</p><h2>O que lembrar nessa data?</h2><p>O 19 de marco pode ser usado por escolas, pesquisadores e moradores como ponto de partida para revisitar personagens, documentos e lugares ligados a formacao de Inhumas.</p><h2>Fontes consultadas</h2><ul><li><a href="https://tede2.pucgoias.edu.br/bitstream/tede/2266/1/EBER%20DORNELAS%20DA%20COSTA%20SOUZA.pdf">Dissertacao PUC Goias - Eber Dornelas da Costa Souza</a></li><li><a href="https://www.ibge.gov.br/cidades-e-estados/go/inhumas.html">IBGE Cidades e Estados - Inhumas GO</a></li></ul>',
  '',
  (SELECT id FROM categories WHERE slug = 'historia-de-inhumas'),
  NULL,
  'draft',
  false,
  'Texto de memoria historica. Revisar calendario local antes de publicar como data comemorativa.',
  'Redacao Inhumas em Foco',
  NULL,
  NULL
);
