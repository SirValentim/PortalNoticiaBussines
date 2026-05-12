INSERT OR IGNORE INTO categories (slug, name, description, requires_editorial_notes)
VALUES
('guia-de-inhumas', 'Guia de Inhumas', 'Informacoes uteis, dados locais e servicos sobre Inhumas GO.', false),
('economia-local', 'Economia Local', 'Comercio, agricultura, pecuaria e desenvolvimento economico de Inhumas.', false),
('meio-ambiente', 'Meio Ambiente', 'Rios, recursos naturais, saneamento e temas ambientais de Inhumas.', false),
('educacao', 'Educacao', 'Escolas, ensino tecnico, universidades e oportunidades de formacao em Inhumas.', false);

INSERT INTO posts (
  title, slug, excerpt, content, cover_image_key, meta_title, meta_description, seo_keyword,
  category_id, author_id, status, is_sponsored, is_featured, editorial_notes, editor_responsible,
  published_at, publish_at
)
VALUES
(
  'Rio Meia-Ponte: por que ele e importante para Inhumas?',
  'rio-meia-ponte-importancia-para-inhumas',
  'O Rio Meia-Ponte aparece em fontes oficiais como o principal rio de Inhumas e tem relacao direta com abastecimento, meio ambiente e planejamento regional.',
  '<p>O Rio Meia-Ponte e um dos elementos naturais mais importantes para entender Inhumas. A propria Prefeitura informa que ele e o principal rio do municipio e responsavel pelo abastecimento de agua da cidade.</p><h2>Onde entra Inhumas nessa bacia?</h2><p>Segundo a pagina de estatisticas da Prefeitura, o Rio Meia-Ponte nasce na Serra do Catingueiro Grande, em Itaucu, e faz parte da hidrografia local. A mesma fonte cita tambem os ribeiroes Peixe, Capoeirao, Inhumas, Quilombo e Lajes.</p><h2>Importancia regional</h2><p>A Secretaria de Meio Ambiente e Desenvolvimento Sustentavel de Goias destaca que a Bacia Hidrografica do Rio Meia Ponte e importante para a Regiao Metropolitana de Goiania e envolve atividades industriais, agroindustriais, pecuaria e producao de hortifrutigranjeiros.</p><h2>Abastecimento e periodo seco</h2><p>Documentos da Saneago sobre abastecimento em Inhumas tambem citam o Rio Meia Ponte como manancial superficial usado no municipio. Isso mostra por que acompanhar chuva, estiagem e conservacao ambiental e relevante para moradores e empresas.</p><h2>Fontes consultadas</h2><ul><li><a href="https://inhumas.go.gov.br/estatisticas/">Prefeitura de Inhumas - Estatisticas</a></li><li><a href="https://goias.gov.br/meioambiente/segurancahidrica/">Semad Goias - Seguranca Hidrica: Rio Meia Ponte</a></li><li><a href="https://www.saneago.com.br/racionamento/Plano_Racionamento_inhumas_2019.pdf">Saneago - Plano de racionamento do abastecimento de agua em Inhumas</a></li></ul>',
  '',
  'Rio Meia-Ponte e Inhumas: importancia para agua e meio ambiente',
  'Entenda por que o Rio Meia-Ponte e importante para Inhumas, com dados da Prefeitura, Semad Goias e Saneago.',
  'Rio Meia-Ponte Inhumas',
  (SELECT id FROM categories WHERE slug = 'meio-ambiente'),
  NULL, 'published', false, false, '', 'Redacao Inhumas em Foco',
  CURRENT_TIMESTAMP, NULL
),
(
  'Economia de Inhumas: agricultura, pecuaria e comercio local',
  'economia-de-inhumas-agricultura-pecuaria-e-comercio-local',
  'Fontes oficiais apontam agricultura, pecuaria e atividades comerciais como partes importantes da economia de Inhumas.',
  '<p>A economia de Inhumas combina atividades rurais, comercio local, servicos e sua posicao regional proxima a Goiania. Para quem empreende ou pesquisa a cidade, entender esse perfil ajuda a enxergar oportunidades.</p><h2>Agricultura e pecuaria</h2><p>A Prefeitura de Inhumas afirma em sua pagina de estatisticas que a terra fertil favorece a agricultura e que a economia local tem relacao com atividades rurais. A pagina tambem apresenta historico de efetivo da pecuaria, producao de leite e ovos.</p><h2>Producao regional</h2><p>A Sudeco, em pagina sobre municipios goianos, resume a economia de Inhumas como movida pela agricultura e pecuaria, com producao de alho, arroz, feijao, milho e hortalicas em geral.</p><h2>Comercio e servicos</h2><p>Alem do campo, Inhumas tem papel comercial para moradores da cidade e da regiao. O guia de lojas, promocoes e servicos do portal deve refletir esse movimento e ajudar empresas locais a serem encontradas.</p><h2>Por que isso importa para o portal?</h2><p>Conteudos sobre economia local ajudam moradores, comerciantes e visitantes a entender o perfil da cidade. Tambem fortalecem buscas como comercio em Inhumas, lojas em Inhumas e economia de Inhumas.</p><h2>Fontes consultadas</h2><ul><li><a href="https://inhumas.go.gov.br/estatisticas/">Prefeitura de Inhumas - Estatisticas</a></li><li><a href="https://www.gov.br/sudeco/pt-br/municipios/goias/marco/inhumas-go">Sudeco - Inhumas GO</a></li><li><a href="https://www.ibge.gov.br/cidades-e-estados/go/inhumas.html">IBGE Cidades e Estados - Inhumas GO</a></li></ul>',
  '',
  'Economia de Inhumas GO: agricultura, pecuaria e comercio',
  'Veja um resumo da economia de Inhumas GO, com agricultura, pecuaria, comercio local e fontes oficiais.',
  'economia de Inhumas',
  (SELECT id FROM categories WHERE slug = 'economia-local'),
  NULL, 'published', false, false, '', 'Redacao Inhumas em Foco',
  CURRENT_TIMESTAMP, NULL
),
(
  'Quais municipios fazem limite com Inhumas?',
  'quais-municipios-fazem-limite-com-inhumas',
  'Inhumas faz limite com Itaucu, Caturai, Goianira, Brazabrantes, Damolandia e Aracu, segundo informacoes da Prefeitura.',
  '<p>Saber quais municipios fazem limite com Inhumas ajuda a entender a localizacao da cidade, seus deslocamentos e sua relacao com a regiao metropolitana de Goiania.</p><h2>Limites informados pela Prefeitura</h2><p>De acordo com a pagina de estatisticas da Prefeitura de Inhumas, o municipio limita-se ao norte com Itaucu; ao sul com Caturai e Goianira; a leste com Brazabrantes e Damolandia; e a oeste com Aracu.</p><h2>O que esses limites significam?</h2><p>Limites municipais podem seguir rios, serras, pontes, estradas ou marcos territoriais. Na pratica, eles ajudam a explicar fluxos de trabalho, estudo, comercio e servicos entre cidades vizinhas.</p><h2>Inhumas no contexto goiano</h2><p>O IBGE identifica Inhumas como municipio de Goias e oferece dados oficiais sobre populacao, area territorial, densidade demografica e indicadores economicos. Essas informacoes ajudam a comparar a cidade com outros municipios goianos.</p><h2>Uso para moradores e empresas</h2><p>Para empresas locais, conhecer os municipios vizinhos pode ajudar em entregas, atendimento regional e campanhas comerciais. Para moradores, ajuda a entender caminhos e conexoes do dia a dia.</p><h2>Fontes consultadas</h2><ul><li><a href="https://inhumas.go.gov.br/estatisticas/">Prefeitura de Inhumas - Estatisticas</a></li><li><a href="https://www.ibge.gov.br/cidades-e-estados/go/inhumas.html">IBGE Cidades e Estados - Inhumas GO</a></li></ul>',
  '',
  'Municipios vizinhos de Inhumas GO: limites e localizacao',
  'Veja quais cidades fazem limite com Inhumas GO e entenda a localizacao do municipio na regiao.',
  'municipios vizinhos de Inhumas',
  (SELECT id FROM categories WHERE slug = 'guia-de-inhumas'),
  NULL, 'published', false, false, '', 'Redacao Inhumas em Foco',
  CURRENT_TIMESTAMP, NULL
),
(
  'IFG Campus Inhumas: educacao profissional e presenca regional',
  'ifg-campus-inhumas-educacao-profissional-e-presenca-regional',
  'O IFG Campus Inhumas aparece como uma das referencias de educacao profissional e tecnica da cidade.',
  '<p>A educacao tambem faz parte da identidade de Inhumas. Entre as instituicoes de destaque esta o Instituto Federal de Goias, que mantem o Campus Inhumas e oferece atividades de ensino, extensao e formacao profissional.</p><h2>Presenca na cidade</h2><p>O portal oficial do IFG possui uma area dedicada ao Campus Inhumas, com noticias, comunicados, cursos e atividades da unidade. A presenca do campus amplia as oportunidades de formacao para jovens e trabalhadores da cidade e da regiao.</p><h2>Participacao comunitaria</h2><p>Em noticia publicada pelo IFG sobre o aniversario de 88 anos de Inhumas, a instituicao registrou a participacao de estudantes e servidores em desfile civico-militar, com atividades ligadas a cultura popular afro-brasileira, como ciranda, maculele e ijexa.</p><h2>Por que isso importa?</h2><p>Instituicoes de educacao profissional ajudam a formar mao de obra, incentivar projetos de extensao e aproximar estudantes de demandas reais do municipio. Para Inhumas, isso significa mais oportunidades de qualificacao e conexao regional.</p><h2>Fontes consultadas</h2><ul><li><a href="https://ifg.edu.br/inhumas">IFG - Campus Inhumas</a></li><li><a href="https://www.ifg.edu.br/estudenoifg/167-ifg/campus/inhumas/noticias-campus-inhumas/12174-inhumas-completa-88-anos-e-ifg-participa-de-desfile-em-homenagem-ao-municipio">IFG - Inhumas completa 88 anos e IFG participa de desfile</a></li></ul>',
  '',
  'IFG Campus Inhumas: educacao profissional na cidade',
  'Conheca a importancia do IFG Campus Inhumas para educacao profissional, extensao e formacao regional.',
  'IFG Campus Inhumas',
  (SELECT id FROM categories WHERE slug = 'educacao'),
  NULL, 'published', false, false, '', 'Redacao Inhumas em Foco',
  CURRENT_TIMESTAMP, NULL
),
(
  'Hidrografia de Inhumas: rios, ribeiroes e recursos naturais',
  'hidrografia-de-inhumas-rios-ribeiroes-e-recursos-naturais',
  'Prefeitura lista o Rio Meia-Ponte e ribeiroes como Peixe, Capoeirao, Inhumas, Quilombo e Lajes na hidrografia local.',
  '<p>A hidrografia de Inhumas e parte importante da paisagem e da vida cotidiana do municipio. Ela influencia abastecimento, meio ambiente, agricultura e planejamento urbano.</p><h2>Rio principal e ribeiroes</h2><p>Segundo a Prefeitura, o Rio Meia-Ponte e o principal rio de Inhumas. A mesma pagina cita ribeiroes como Peixe, Capoeirao, Inhumas, Quilombo e Lajes.</p><h2>Recursos naturais</h2><p>A Prefeitura tambem destaca a fertilidade da terra e a presenca de vegetacao e matas como recursos naturais do municipio. Esses elementos ajudam a explicar a relacao historica da cidade com agricultura e pecuaria.</p><h2>Conservacao e seguranca hidrica</h2><p>A Semad Goias afirma que a bacia do Rio Meia Ponte exige governanca das aguas por envolver multiplos usos, incluindo atividades agroindustriais, pecuaria e producao de hortifrutigranjeiros.</p><h2>Informacao util para a cidade</h2><p>Quando o morador entende os rios e ribeiroes do municipio, tambem entende melhor temas como chuva, estiagem, abastecimento, drenagem e preservacao ambiental.</p><h2>Fontes consultadas</h2><ul><li><a href="https://inhumas.go.gov.br/estatisticas/">Prefeitura de Inhumas - Estatisticas</a></li><li><a href="https://goias.gov.br/meioambiente/segurancahidrica/">Semad Goias - Seguranca Hidrica</a></li></ul>',
  '',
  'Hidrografia de Inhumas GO: rios e ribeiroes',
  'Conheca a hidrografia de Inhumas GO, incluindo Rio Meia-Ponte, ribeiroes e recursos naturais.',
  'hidrografia de Inhumas',
  (SELECT id FROM categories WHERE slug = 'meio-ambiente'),
  NULL, 'published', false, false, '', 'Redacao Inhumas em Foco',
  CURRENT_TIMESTAMP, NULL
),
(
  'Guia rapido de Inhumas GO: populacao, localizacao e perfil da cidade',
  'guia-rapido-de-inhumas-go-populacao-localizacao-e-perfil',
  'Um resumo pratico com dados de populacao, territorio, economia e localizacao de Inhumas em Goias.',
  '<p>Inhumas e um municipio goiano com forte relacao regional, atividade comercial, base rural e conexao com a Regiao Metropolitana de Goiania. Para quem chegou agora ou pesquisa a cidade, alguns dados ajudam a formar uma visao geral.</p><h2>Populacao e territorio</h2><p>O IBGE Cidades e Estados apresenta dados oficiais de populacao, area territorial, densidade demografica, economia e outros indicadores de Inhumas. Esses numeros sao usados como referencia por gestores, empresas, pesquisadores e moradores.</p><h2>Localizacao e limites</h2><p>A Prefeitura informa que Inhumas faz limite com Itaucu, Caturai, Goianira, Brazabrantes, Damolandia e Aracu. Esses municipios ajudam a formar o entorno regional da cidade.</p><h2>Economia e vocacao local</h2><p>Fontes institucionais apontam agricultura e pecuaria como partes importantes da economia, com producao rural e comercio local compondo o perfil do municipio.</p><h2>Como usar este guia</h2><p>Este resumo serve como ponto de partida para estudantes, moradores, visitantes e empresas que precisam de informacoes basicas sobre Inhumas GO.</p><h2>Fontes consultadas</h2><ul><li><a href="https://www.ibge.gov.br/cidades-e-estados/go/inhumas.html">IBGE Cidades e Estados - Inhumas GO</a></li><li><a href="https://inhumas.go.gov.br/estatisticas/">Prefeitura de Inhumas - Estatisticas</a></li><li><a href="https://www.gov.br/sudeco/pt-br/municipios/goias/marco/inhumas-go">Sudeco - Inhumas GO</a></li></ul>',
  '',
  'Guia rapido de Inhumas GO: populacao e perfil da cidade',
  'Resumo pratico sobre Inhumas GO com populacao, localizacao, economia e fontes oficiais.',
  'Inhumas GO',
  (SELECT id FROM categories WHERE slug = 'guia-de-inhumas'),
  NULL, 'published', false, false, '', 'Redacao Inhumas em Foco',
  CURRENT_TIMESTAMP, NULL
)
ON CONFLICT(slug) DO UPDATE SET
  title=excluded.title,
  excerpt=excluded.excerpt,
  content=excluded.content,
  meta_title=excluded.meta_title,
  meta_description=excluded.meta_description,
  seo_keyword=excluded.seo_keyword,
  category_id=excluded.category_id,
  status='published',
  editor_responsible=excluded.editor_responsible,
  published_at=COALESCE(posts.published_at, CURRENT_TIMESTAMP),
  updated_at=CURRENT_TIMESTAMP;

INSERT INTO jobs (type, payload, status, run_at, max_attempts)
VALUES ('generate_sitemap', '{}', 'pending', CURRENT_TIMESTAMP, 3);
