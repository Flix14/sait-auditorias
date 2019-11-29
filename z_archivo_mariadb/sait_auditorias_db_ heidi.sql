-- --------------------------------------------------------
-- Host:                         127.0.0.1
-- Versión del servidor:         10.4.8-MariaDB - mariadb.org binary distribution
-- SO del servidor:              Win64
-- HeidiSQL Versión:             10.2.0.5599
-- --------------------------------------------------------

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET NAMES utf8 */;
/*!50503 SET NAMES utf8mb4 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;


-- Volcando estructura de base de datos para saitauditorias
CREATE DATABASE IF NOT EXISTS `saitauditorias` /*!40100 DEFAULT CHARACTER SET utf8 */;
USE `saitauditorias`;

-- Volcando estructura para tabla saitauditorias.auditorias
CREATE TABLE IF NOT EXISTS `auditorias` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `motivo` varchar(50) NOT NULL,
  `comentario` text DEFAULT NULL,
  `comandos` mediumtext NOT NULL,
  `fecha` datetime NOT NULL DEFAULT current_timestamp(),
  `id_usuario` int(10) unsigned NOT NULL,
  `id_proyecto_servidor` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_usuarios` (`id_usuario`),
  KEY `fk_proyectos_servidores` (`id_proyecto_servidor`),
  FULLTEXT KEY `motivo` (`motivo`),
  CONSTRAINT `fk_proyectos_servidores` FOREIGN KEY (`id_proyecto_servidor`) REFERENCES `proyectos_servidores` (`id`),
  CONSTRAINT `fk_usuarios` FOREIGN KEY (`id_usuario`) REFERENCES `usuarios` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=54 DEFAULT CHARSET=utf8;

-- La exportación de datos fue deseleccionada.

-- Volcando estructura para tabla saitauditorias.dominios
CREATE TABLE IF NOT EXISTS `dominios` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `id_servidor` int(10) unsigned NOT NULL,
  `dominio` varchar(50) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `dominio` (`dominio`),
  KEY `fk_dominios_servidores` (`id_servidor`),
  CONSTRAINT `fk_dominios_servidores` FOREIGN KEY (`id_servidor`) REFERENCES `servidores` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=199 DEFAULT CHARSET=utf8;

-- La exportación de datos fue deseleccionada.

-- Volcando estructura para tabla saitauditorias.proyectos
CREATE TABLE IF NOT EXISTS `proyectos` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `nombre` varchar(50) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `nombre` (`nombre`)
) ENGINE=InnoDB AUTO_INCREMENT=43 DEFAULT CHARSET=utf8;

-- La exportación de datos fue deseleccionada.

-- Volcando estructura para tabla saitauditorias.proyectos_servidores
CREATE TABLE IF NOT EXISTS `proyectos_servidores` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `id_proyecto` int(10) unsigned NOT NULL,
  `id_servidor` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_proyectos` (`id_proyecto`),
  KEY `fk_servidores` (`id_servidor`),
  CONSTRAINT `fk_proyectos` FOREIGN KEY (`id_proyecto`) REFERENCES `proyectos` (`id`),
  CONSTRAINT `fk_servidores` FOREIGN KEY (`id_servidor`) REFERENCES `servidores` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=75 DEFAULT CHARSET=utf8;

-- La exportación de datos fue deseleccionada.

-- Volcando estructura para tabla saitauditorias.servidores
CREATE TABLE IF NOT EXISTS `servidores` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `direccion_publica` varchar(15) NOT NULL,
  `sistema_operativo` varchar(50) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `direccion_publica` (`direccion_publica`)
) ENGINE=InnoDB AUTO_INCREMENT=63 DEFAULT CHARSET=utf8;

-- La exportación de datos fue deseleccionada.

-- Volcando estructura para tabla saitauditorias.usuarios
CREATE TABLE IF NOT EXISTS `usuarios` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `estado` tinyint(4) NOT NULL DEFAULT 1,
  `email` varchar(150) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `email` (`email`)
) ENGINE=InnoDB AUTO_INCREMENT=86 DEFAULT CHARSET=utf8;

-- La exportación de datos fue deseleccionada.

/*!40101 SET SQL_MODE=IFNULL(@OLD_SQL_MODE, '') */;
/*!40014 SET FOREIGN_KEY_CHECKS=IF(@OLD_FOREIGN_KEY_CHECKS IS NULL, 1, @OLD_FOREIGN_KEY_CHECKS) */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
