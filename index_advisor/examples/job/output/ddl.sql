CREATE INDEX idx_movie_id_person_id ON imdbload_no_fk.cast_info (movie_id, person_id);
CREATE INDEX idx_person_id ON imdbload_no_fk.cast_info (person_id);
CREATE INDEX idx_role_id ON imdbload_no_fk.cast_info (role_id);
CREATE INDEX idx_movie_id_company_id_company_type_id ON imdbload_no_fk.movie_companies (movie_id, company_id, company_type_id);
CREATE INDEX idx_info_type_id ON imdbload_no_fk.movie_info (info_type_id);
CREATE INDEX idx_movie_id_info_type_id ON imdbload_no_fk.movie_info (movie_id, info_type_id);
CREATE INDEX idx_movie_id_info_type_id ON imdbload_no_fk.movie_info_idx (movie_id, info_type_id);
CREATE INDEX idx_keyword_id_movie_id ON imdbload_no_fk.movie_keyword (keyword_id, movie_id);
CREATE INDEX idx_movie_id_keyword_id ON imdbload_no_fk.movie_keyword (movie_id, keyword_id)