from app.ai.ingest import _chunk, _split_sections
from app.ai.vectorstore import _tsquery_or_terms


def test_split_sections_breaks_at_headings():
    text = "preamble\n\n# One\n\nbody one\n\n## Two\n\nbody two"
    sections = _split_sections(text)
    assert len(sections) == 3
    assert sections[0].startswith("preamble")
    assert sections[1].startswith("# One")
    assert sections[2].startswith("## Two")


def test_split_sections_without_headings_is_whole_text():
    assert _split_sections("just\n\nparagraphs") == ["just\n\nparagraphs"]


def test_chunk_never_spans_a_heading_boundary():
    # Both sections are tiny, so pure length-packing would merge them; the
    # heading split must keep the second section's answer in its own chunk.
    text = "# Endpoints\n\n- url one\n\n- url two\n\n# Updating certificates\n\nupload a new metadata file"
    chunks = _chunk(text, max_chars=1200)
    assert len(chunks) == 2
    assert chunks[1] == "# Updating certificates\n\nupload a new metadata file"


def test_chunk_still_packs_and_slices_within_a_section():
    long_para = "x" * 250
    text = f"# S\n\n{'a' * 90}\n\n{'b' * 90}\n\n{long_para}"
    chunks = _chunk(text, max_chars=100)
    assert chunks[0] == "# S\n\n" + "a" * 90
    assert chunks[1] == "b" * 90
    # Oversized paragraph is hard-sliced.
    assert [len(c) for c in chunks[2:]] == [100, 100, 50]


def test_tsquery_or_terms_ors_unique_alphanumeric_tokens():
    q = 'SSO login failing: "Invalid SAML Response" SSO 200+'
    assert _tsquery_or_terms(q) == "sso | login | failing | invalid | saml | response | 200"


def test_tsquery_or_terms_empty_query():
    assert _tsquery_or_terms("") == ""
    assert _tsquery_or_terms("!!! ???") == ""
